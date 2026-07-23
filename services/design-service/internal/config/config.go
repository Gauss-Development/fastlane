package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port                     string
	GRPCPort                 string
	Environment              string
	LogLevel                 string
	Database                 DatabaseConfig
	GRPCTLS                  GRPCTLSConfig
	ServiceTransportSecurity string
	InternalHTTPTrustMode    string
	EnableGRPCReflection     bool
	Storage                  StorageConfig
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // minutes
}

type GRPCTLSConfig struct {
	Enabled           bool
	CAFile            string
	CertFile          string
	KeyFile           string
	RequireClientCert bool
}

type StorageConfig struct {
	Endpoint         string // public/browser-reachable host baked into presigned URLs
	InternalEndpoint string // in-network host for live ops (StatObject) — dialed from the container
	AccessKey        string
	SecretKey        string
	Bucket           string
	UseSSL           bool
	Region           string
	PresignTTL       time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnv("PORT", "8086"),
		GRPCPort:    getEnv("GRPC_PORT", "50056"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Database: DatabaseConfig{
			URL:             os.Getenv("DATABASE_URL"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 60),
		},
		GRPCTLS: GRPCTLSConfig{
			Enabled:           getEnvAsBool("GRPC_TLS_ENABLED", false),
			CAFile:            getEnv("GRPC_TLS_CA_FILE", ""),
			CertFile:          getEnv("GRPC_TLS_CERT_FILE", ""),
			KeyFile:           getEnv("GRPC_TLS_KEY_FILE", ""),
			RequireClientCert: getEnvAsBool("GRPC_TLS_REQUIRE_CLIENT_CERT", false),
		},
		ServiceTransportSecurity: resolveTransportSecurityMode(
			getEnv("SERVICE_TRANSPORT_SECURITY", ""),
			getEnv("ENVIRONMENT", "development"),
			getEnvAsBool("GRPC_TLS_ENABLED", false),
		),
		InternalHTTPTrustMode: resolveInternalHTTPTrustMode(
			getEnv("INTERNAL_HTTP_TRUST_MODE", ""),
			getEnv("ENVIRONMENT", "development"),
		),
		EnableGRPCReflection: getEnvAsBool("GRPC_REFLECTION_ENABLED", getEnv("ENVIRONMENT", "development") != "production"),
		Storage: StorageConfig{
			Endpoint:   getEnv("S3_ENDPOINT", "minio:9000"),
			AccessKey:  os.Getenv("S3_ACCESS_KEY"),
			SecretKey:  os.Getenv("S3_SECRET_KEY"),
			InternalEndpoint: getEnv("S3_INTERNAL_ENDPOINT", "minio:9000"),
			Bucket:     getEnv("S3_BUCKET", "design-files"),
			UseSSL:     getEnvAsBool("S3_USE_SSL", false),
			// Region must be set so presigning stays offline (no GetBucketLocation
			// dial) — S3_ENDPOINT is a browser-reachable host the container can't hit.
			Region:     getEnv("S3_REGION", "us-east-1"),
			PresignTTL: getEnvAsDuration("PRESIGN_TTL", 15*time.Minute),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.GRPCPort == "" {
		return fmt.Errorf("GRPC_PORT is required")
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
	if err := validateInternalHTTPTrustMode(c.Environment, c.InternalHTTPTrustMode); err != nil {
		return err
	}
	if c.Environment == "production" && c.EnableGRPCReflection {
		return fmt.Errorf("GRPC_REFLECTION_ENABLED cannot be true in production")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
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

func resolveInternalHTTPTrustMode(value, environment string) string {
	mode := strings.ToLower(strings.TrimSpace(value))
	if mode != "" {
		return mode
	}
	if environment == "production" {
		return ""
	}
	return "insecure_dev"
}

func validateInternalHTTPTrustMode(environment, mode string) error {
	switch mode {
	case "private_network", "disabled":
		return nil
	case "insecure_dev":
		if environment == "production" {
			return fmt.Errorf("INTERNAL_HTTP_TRUST_MODE=insecure_dev is not allowed in production")
		}
		return nil
	case "":
		if environment == "production" {
			return fmt.Errorf("INTERNAL_HTTP_TRUST_MODE is required in production")
		}
		return nil
	default:
		return fmt.Errorf("INTERNAL_HTTP_TRUST_MODE must be one of private_network, disabled, insecure_dev")
	}
}
