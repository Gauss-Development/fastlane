package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	searchv1 "github.com/nikitashilov/microblog_grpc/proto/search/v1"
	"search-service/internal/application/services"
	"search-service/internal/config"
	"search-service/internal/infrastructure/anthropic"
	"search-service/internal/infrastructure/embedding"
	"search-service/internal/infrastructure/postgres"
	"search-service/internal/infrastructure/redis"
	grpcinterface "search-service/internal/interfaces/grpc"
	"search-service/pkg/logger"
	"search-service/pkg/metrics"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	grpc_reflection "google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	metrics.Init()
	appLogger := logger.New(cfg.LogLevel)

	// Catalog DB (read-only): products + embeddings + suppliers, owned by post-service.
	db, err := postgres.NewConnection(postgres.Config{
		URL:             cfg.Catalog.URL,
		MaxOpenConns:    cfg.Catalog.MaxOpenConns,
		MaxIdleConns:    cfg.Catalog.MaxIdleConns,
		ConnMaxLifetime: cfg.Catalog.ConnMaxLifetime,
	})
	if err != nil {
		appLogger.Fatal("catalog DB: " + err.Error())
	}
	defer db.Close()
	catalogRepo := postgres.NewCatalogRepo(db)

	embedder := newEmbedder(cfg.AI, appLogger)
	claude := anthropic.NewClient(cfg.AI.AnthropicAPIKey, cfg.AI.AnthropicModel)
	if !claude.Enabled() {
		appLogger.Warn("ANTHROPIC_API_KEY not set; spec extraction and match explanations are disabled (vector-only results)")
	}

	var cache services.Cache = services.NoopCache{}
	var redisCache *redis.Cache
	if cfg.Redis.Enabled {
		redisCache = redis.NewCache(cfg.Redis.URL, cfg.Redis.Password, cfg.Redis.DB)
		pingCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		if err := redisCache.Ping(pingCtx); err != nil {
			appLogger.Warn("redis ping failed; running without cache: " + err.Error())
			_ = redisCache.Close()
			redisCache = nil
		} else {
			cache = redisCache
		}
		cancel()
	} else {
		appLogger.Info("REDIS_URL not set; running without AI-step cache")
	}
	if redisCache != nil {
		defer redisCache.Close()
	}

	searchSvc := services.NewSearchService(claude, embedder, catalogRepo, claude, cache, cfg.CacheTTL, appLogger)

	grpcOptions := []grpc.ServerOption{
		// Match the keepalive enforcement of the other gRPC services. Without
		// this the server uses gRPC's default policy (MinTime 5m,
		// PermitWithoutStream=false) and GOAWAYs the api-gateway client — which
		// pings every 30s with PermitWithoutStream — with "too_many_pings".
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Minute,
			MaxConnectionAge:      30 * time.Minute,
			MaxConnectionAgeGrace: 5 * time.Minute,
			Time:                  5 * time.Second,
			Timeout:               1 * time.Second,
		}),
		grpc.ChainUnaryInterceptor(
			metrics.UnaryServerInterceptor("search-service"),
			unaryLoggingInterceptor(appLogger),
		),
	}
	if cfg.GRPCTLS.Enabled {
		transportCreds, credsErr := buildServerTransportCredentials(cfg.GRPCTLS)
		if credsErr != nil {
			appLogger.Fatal("Failed to configure gRPC TLS credentials: " + credsErr.Error())
		}
		grpcOptions = append(grpcOptions, grpc.Creds(transportCreds))
	}

	grpcServer := grpc.NewServer(grpcOptions...)
	searchv1.RegisterSearchServiceServer(grpcServer, grpcinterface.NewSearchServer(searchSvc, appLogger))
	if cfg.EnableGRPCReflection {
		grpc_reflection.Register(grpcServer)
	}

	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		appLogger.Fatal("gRPC listen: " + err.Error())
	}

	go func() {
		appLogger.Info("Search gRPC server listening on :" + cfg.GRPCPort)
		if err := grpcServer.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			appLogger.Fatal("gRPC serve: " + err.Error())
		}
	}()

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", metrics.Handler())
	metricsMux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	metricsSrv := &http.Server{
		Addr:              ":" + cfg.MetricsHTTPPort,
		Handler:           metricsMux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		appLogger.Info("Search metrics/health HTTP listening on :" + cfg.MetricsHTTPPort)
		if err := metricsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Fatal("metrics HTTP: " + err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := metricsSrv.Shutdown(shutdownCtx); err != nil {
		appLogger.Warn("metrics HTTP shutdown: " + err.Error())
	}

	grpcServer.GracefulStop()
	appLogger.Info("Done")
}

// newEmbedder picks the embedding provider by available key: Voyage (primary),
// then OpenAI (fallback), then a deterministic fake so the service still runs
// locally without keys. Must match the provider used at seed time for vectors
// to be comparable.
func newEmbedder(ai config.AIConfig, log *logger.Logger) embedding.Client {
	switch {
	case ai.VoyageAPIKey != "":
		log.Info("embeddings: voyage-3")
		return embedding.NewVoyageClient(ai.VoyageAPIKey)
	case ai.OpenAIAPIKey != "":
		log.Info("embeddings: openai text-embedding-3-small")
		return embedding.NewOpenAIClient(ai.OpenAIAPIKey)
	default:
		log.Warn("no embedding API key set; using deterministic fake embedder (search results will not be semantically meaningful)")
		return embedding.NewFakeClient()
	}
}

func unaryLoggingInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		log.Debug(info.FullMethod + " " + time.Since(start).String())
		return resp, err
	}
}

func buildServerTransportCredentials(tlsCfg config.GRPCTLSConfig) (credentials.TransportCredentials, error) {
	serverCert, err := tls.LoadX509KeyPair(tlsCfg.CertFile, tlsCfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load gRPC server certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{serverCert},
	}

	if tlsCfg.RequireClientCert {
		caPEM, caErr := os.ReadFile(tlsCfg.CAFile)
		if caErr != nil {
			return nil, fmt.Errorf("read gRPC CA file: %w", caErr)
		}

		clientCAs := x509.NewCertPool()
		if ok := clientCAs.AppendCertsFromPEM(caPEM); !ok {
			return nil, fmt.Errorf("parse gRPC client CA certificate")
		}

		tlsConfig.ClientCAs = clientCAs
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return credentials.NewTLS(tlsConfig), nil
}
