package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"order-service/internal/application/services"
	"order-service/internal/config"
	"order-service/internal/infrastructure/messaging"
	"order-service/internal/infrastructure/postgres"
	grpcinterface "order-service/internal/interfaces/grpc"
	"order-service/pkg/logger"
	"order-service/pkg/metrics"

	orderv1 "github.com/nikitashilov/microblog_grpc/proto/order/v1"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	grpc_reflection "google.golang.org/grpc/reflection"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	appLogger := logger.New(cfg.LogLevel)
	metrics.Init()

	appLogger.Info("transport security: " + cfg.ServiceTransportSecurity)
	appLogger.Info("internal HTTP trust: " + cfg.InternalHTTPTrustMode)

	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		appLogger.Fatal("Failed to connect to database: " + err.Error())
	}
	defer db.Close()

	if err := postgres.RunMigrations(db, migrationsFS); err != nil {
		appLogger.Fatal("Failed to run migrations: " + err.Error())
	}

	orderRepo := postgres.NewOrderRepository(db)
	eventRepo := postgres.NewOrderEventRepository(db)
	orderSvc := services.NewOrderService(orderRepo, eventRepo, appLogger)

	// Start RabbitMQ consumer (no-op if RABBITMQ_URL is empty).
	consumer := messaging.NewConsumer(cfg.RabbitMQ, orderSvc, appLogger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := consumer.Start(ctx); err != nil {
		appLogger.Fatal("Failed to start order consumer: " + err.Error())
	}
	defer consumer.Close()

	grpcOptions := []grpc.ServerOption{
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
			metrics.UnaryServerInterceptor("order-service"),
			unaryServerLoggingInterceptor(appLogger),
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
	orderv1.RegisterOrderServiceServer(grpcServer, grpcinterface.NewOrderServer(orderSvc, appLogger))
	if cfg.EnableGRPCReflection {
		grpc_reflection.Register(grpcServer)
	}

	grpcListener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		appLogger.Fatal("Failed to create gRPC listener: " + err.Error())
	}

	go func() {
		appLogger.Info("Order gRPC service starting on port " + cfg.GRPCPort)
		if serveErr := grpcServer.Serve(grpcListener); serveErr != nil && !errors.Is(serveErr, grpc.ErrServerStopped) {
			appLogger.Fatal("Failed to start gRPC server: " + serveErr.Error())
		}
	}()

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(metrics.GinMiddleware("order-service"))
	router.GET("/metrics", gin.WrapH(metrics.Handler()))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		appLogger.Info("Order service starting on port " + cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start server: " + err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutCancel()

	if err := server.Shutdown(shutCtx); err != nil {
		appLogger.Fatal("HTTP server forced to shutdown: " + err.Error())
	}

	grpcServer.GracefulStop()
	appLogger.Info("Servers exited")
}

func unaryServerLoggingInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)
		if err != nil {
			log.Warn(fmt.Sprintf("gRPC method %s failed: %v (duration: %v)", info.FullMethod, err, duration))
		} else {
			log.Debug(fmt.Sprintf("gRPC method %s succeeded (duration: %v)", info.FullMethod, duration))
		}
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
