package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	GRPCPort                 string
	MetricsHTTPPort          string
	Environment              string
	LogLevel                 string
	Catalog                  CatalogConfig
	Redis                    RedisConfig
	AI                       AIConfig
	CacheTTL                 time.Duration
	GRPCTLS                  GRPCTLSConfig
	ServiceTransportSecurity string
	EnableGRPCReflection     bool
}

// CatalogConfig points at the post-service-owned catalog database. search-service
// reads it (products + embeddings + suppliers) read-only for vector ranking.
type CatalogConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // minutes
}

type RedisConfig struct {
	URL      string
	Password string
	DB       int
	Enabled  bool
}

// AIConfig holds the external AI provider keys. Empty/placeholder keys put the
// service in a degraded mode: a missing embedding key falls back to the
// deterministic fake embedder; a missing Anthropic key skips spec extraction
// and match explanations (vector-only results).
type AIConfig struct {
	AnthropicAPIKey string
	AnthropicModel  string
	VoyageAPIKey    string
	OpenAIAPIKey    string
}

type GRPCTLSConfig struct {
	Enabled           bool
	CAFile            string
	CertFile          string
	KeyFile           string
	RequireClientCert bool
}

func Load() (*Config, error) {
	cfg := &Config{
		GRPCPort:        getEnv("GRPC_PORT", "50054"),
		MetricsHTTPPort: getEnv("METRICS_HTTP_PORT", "9095"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		Catalog: CatalogConfig{
			URL:             getEnv("CATALOG_DATABASE_URL", ""),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 60),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", ""),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			Enabled:  getEnv("REDIS_URL", "") != "",
		},
		AI: AIConfig{
			AnthropicAPIKey: realKey(getEnv("ANTHROPIC_API_KEY", "")),
			AnthropicModel:  getEnv("ANTHROPIC_MODEL", "claude-haiku-4-5"),
			VoyageAPIKey:    realKey(getEnv("VOYAGE_API_KEY", "")),
			OpenAIAPIKey:    realKey(getEnv("OPENAI_API_KEY", "")),
		},
		CacheTTL: time.Duration(getEnvAsInt("CACHE_TTL_HOURS", 24)) * time.Hour,
		GRPCTLS: GRPCTLSConfig{
			Enabled:           getEnvAsBool("GRPC_TLS_ENABLED", false),
			CAFile:            getEnv("GRPC_TLS_CA_FILE", ""),
			CertFile:          getEnv("GRPC_TLS_CERT_FILE", ""),
			KeyFile:           getEnv("GRPC_TLS_KEY_FILE", ""),
			RequireClientCert: getEnvAsBool("GRPC_TLS_REQUIRE_CLIENT_CERT", false),
		},
		ServiceTransportSecurity: resolveTransportSecurityMode(getEnv("SERVICE_TRANSPORT_SECURITY", ""), getEnv("ENVIRONMENT", "development"), getEnvAsBool("GRPC_TLS_ENABLED", false)),
		EnableGRPCReflection:     getEnvAsBool("GRPC_REFLECTION_ENABLED", getEnv("ENVIRONMENT", "development") != "production"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// realKey treats empty or ".env.example" placeholder values as unset, so a
// freshly-cloned environment runs in degraded mode rather than sending a
// placeholder string to a provider.
func realKey(v string) string {
	if v == "" || strings.HasPrefix(v, "replace-with") {
		return ""
	}
	return v
}

func (c *Config) validate() error {
	if c.GRPCPort == "" {
		return fmt.Errorf("GRPC_PORT is required")
	}
	if c.MetricsHTTPPort == "" {
		return fmt.Errorf("METRICS_HTTP_PORT is required")
	}
	if c.MetricsHTTPPort == c.GRPCPort {
		return fmt.Errorf("METRICS_HTTP_PORT must differ from GRPC_PORT")
	}
	if c.Catalog.URL == "" {
		return fmt.Errorf("CATALOG_DATABASE_URL is required")
	}
	if c.GRPCTLS.Enabled {
		if c.GRPCTLS.CAFile == "" {
			return fmt.Errorf("GRPC_TLS_CA_FILE is required when GRPC_TLS_ENABLED=true")
		}
		if c.GRPCTLS.CertFile == "" || c.GRPCTLS.KeyFile == "" {
			return fmt.Errorf("GRPC_TLS_CERT_FILE and GRPC_TLS_KEY_FILE are required when GRPC_TLS_ENABLED=true")
		}
	}
	if (c.GRPCTLS.CertFile == "") != (c.GRPCTLS.KeyFile == "") {
		return fmt.Errorf("GRPC_TLS_CERT_FILE and GRPC_TLS_KEY_FILE must be set together")
	}
	if err := validateTransportSecurityMode(c.Environment, c.ServiceTransportSecurity, c.GRPCTLS.Enabled); err != nil {
		return err
	}
	if c.Environment == "production" && c.EnableGRPCReflection {
		return fmt.Errorf("GRPC_REFLECTION_ENABLED cannot be true in production")
	}
	return nil
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func resolveTransportSecurityMode(value, environment string, grpcTLSEnabled bool) string {
	mode := strings.ToLower(strings.TrimSpace(value))
	if mode != "" {
		return mode
	}
	if environment == "production" {
		return ""
	}
	if grpcTLSEnabled {
		return "app_mtls"
	}
	return "insecure_dev"
}

func validateTransportSecurityMode(environment, mode string, grpcTLSEnabled bool) error {
	switch mode {
	case "mesh":
		return nil
	case "app_mtls":
		if !grpcTLSEnabled {
			return fmt.Errorf("GRPC_TLS_ENABLED=true is required when SERVICE_TRANSPORT_SECURITY=app_mtls")
		}
		return nil
	case "insecure_dev":
		if environment == "production" {
			return fmt.Errorf("SERVICE_TRANSPORT_SECURITY=insecure_dev is not allowed in production")
		}
		return nil
	case "":
		if environment == "production" {
			return fmt.Errorf("SERVICE_TRANSPORT_SECURITY is required in production")
		}
		return nil
	default:
		return fmt.Errorf("SERVICE_TRANSPORT_SECURITY must be one of mesh, app_mtls, insecure_dev")
	}
}
