package clients

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"post-service/internal/config"

	authv1 "github.com/nikitashilov/microblog_grpc/proto/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

const defaultAuthTimeout = 10 * time.Second

// AuthClient mints supplier magic-link tokens via auth-service, which owns
// the magic-link signing key (separate from access/refresh tokens).
type AuthClient struct {
	conn   *grpc.ClientConn
	client authv1.AuthServiceClient
}

func NewAuthClient(addr string, tlsCfg config.GRPCTLSConfig) (*AuthClient, error) {
	creds, err := buildClientTransportCredentials(tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("build auth client transport credentials: %w", err)
	}

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(creds),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to auth gRPC service: %w", err)
	}

	return &AuthClient{
		conn:   conn,
		client: authv1.NewAuthServiceClient(conn),
	}, nil
}

func (c *AuthClient) IssueMagicLinkToken(ctx context.Context, rfqID, supplierID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultAuthTimeout)
	defer cancel()

	resp, err := c.client.IssueMagicLinkToken(ctx, &authv1.IssueMagicLinkTokenRequest{
		RfqId:      rfqID,
		SupplierId: supplierID,
	})
	if err != nil {
		return "", err
	}
	return resp.GetToken(), nil
}

func (c *AuthClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func buildClientTransportCredentials(tlsCfg config.GRPCTLSConfig) (credentials.TransportCredentials, error) {
	if !tlsCfg.Enabled {
		return insecure.NewCredentials(), nil
	}

	caPEM, err := os.ReadFile(tlsCfg.CAFile)
	if err != nil {
		return nil, fmt.Errorf("read gRPC CA file: %w", err)
	}

	rootCAs := x509.NewCertPool()
	if ok := rootCAs.AppendCertsFromPEM(caPEM); !ok {
		return nil, fmt.Errorf("parse gRPC CA certificate")
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    rootCAs,
	}

	if tlsCfg.CertFile != "" && tlsCfg.KeyFile != "" {
		clientCert, certErr := tls.LoadX509KeyPair(tlsCfg.CertFile, tlsCfg.KeyFile)
		if certErr != nil {
			return nil, fmt.Errorf("load gRPC client certificate: %w", certErr)
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
	}

	return credentials.NewTLS(tlsConfig), nil
}
