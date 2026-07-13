package clients

import (
	"context"
	"fmt"
	"time"

	"api-gateway/internal/config"
	"api-gateway/pkg/logger"

	manufacturerv1 "github.com/nikitashilov/microblog_grpc/proto/manufacturer/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/types/known/emptypb"
)

const defaultManufacturerTimeout = 15 * time.Second

// ManufacturerClient talks to the ManufacturerService hosted by catalog-service
// (Chinese PCB/PCBA manufacturer profiles + capabilities).
type ManufacturerClient struct {
	conn   *grpc.ClientConn
	client manufacturerv1.ManufacturerServiceClient
	logger *logger.Logger
}

func NewManufacturerClient(addr string, tlsCfg config.GRPCTLSConfig, logger *logger.Logger) (*ManufacturerClient, error) {
	creds, err := buildClientTransportCredentials(tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("build manufacturer client transport credentials: %w", err)
	}

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(creds),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                keepaliveTime,
			Timeout:             keepaliveTimeout,
			PermitWithoutStream: keepalivePermitWithoutStream,
		}),
		grpc.WithUnaryInterceptor(unaryClientLoggingInterceptor(logger)),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to catalog gRPC service: %w", err)
	}
	return &ManufacturerClient{
		conn:   conn,
		client: manufacturerv1.NewManufacturerServiceClient(conn),
		logger: logger,
	}, nil
}

func (c *ManufacturerClient) CreateManufacturer(ctx context.Context, m *manufacturerv1.Manufacturer) (*manufacturerv1.Manufacturer, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultManufacturerTimeout)
	defer cancel()
	return c.client.CreateManufacturer(ctx, &manufacturerv1.CreateManufacturerRequest{Manufacturer: m})
}

func (c *ManufacturerClient) GetManufacturer(ctx context.Context, id string) (*manufacturerv1.Manufacturer, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultManufacturerTimeout)
	defer cancel()
	return c.client.GetManufacturer(ctx, &manufacturerv1.GetManufacturerRequest{Id: id})
}

func (c *ManufacturerClient) GetManufacturerByUser(ctx context.Context, userID string) (*manufacturerv1.Manufacturer, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultManufacturerTimeout)
	defer cancel()
	return c.client.GetManufacturerByUser(ctx, &manufacturerv1.GetManufacturerByUserRequest{UserId: userID})
}

func (c *ManufacturerClient) UpdateManufacturer(ctx context.Context, m *manufacturerv1.Manufacturer, actorID string) (*manufacturerv1.Manufacturer, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultManufacturerTimeout)
	defer cancel()
	return c.client.UpdateManufacturer(ctx, &manufacturerv1.UpdateManufacturerRequest{Manufacturer: m, ActorId: actorID})
}

func (c *ManufacturerClient) ListManufacturers(ctx context.Context, req *manufacturerv1.ListManufacturersRequest) (*manufacturerv1.ListManufacturersResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultManufacturerTimeout)
	defer cancel()
	return c.client.ListManufacturers(ctx, req)
}

func (c *ManufacturerClient) VerifyManufacturer(ctx context.Context, id string, verified bool, actorID string) (*manufacturerv1.Manufacturer, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultManufacturerTimeout)
	defer cancel()
	return c.client.VerifyManufacturer(ctx, &manufacturerv1.VerifyManufacturerRequest{Id: id, Verified: verified, ActorId: actorID})
}

func (c *ManufacturerClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := c.client.HealthCheck(ctx, &emptypb.Empty{})
	return err
}

func (c *ManufacturerClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
