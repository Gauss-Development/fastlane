package clients

import (
	"context"
	"fmt"
	"time"

	"api-gateway/internal/config"
	"api-gateway/pkg/logger"

	orderv1 "github.com/nikitashilov/microblog_grpc/proto/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/types/known/emptypb"
)

const defaultOrderTimeout = 15 * time.Second

// OrderClient talks to the OrderService.
type OrderClient struct {
	conn   *grpc.ClientConn
	client orderv1.OrderServiceClient
	logger *logger.Logger
}

func NewOrderClient(addr string, tlsCfg config.GRPCTLSConfig, logger *logger.Logger) (*OrderClient, error) {
	creds, err := buildClientTransportCredentials(tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("build order client transport credentials: %w", err)
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
		return nil, fmt.Errorf("connect to order gRPC service: %w", err)
	}
	return &OrderClient{
		conn:   conn,
		client: orderv1.NewOrderServiceClient(conn),
		logger: logger,
	}, nil
}

func (c *OrderClient) ListOrders(ctx context.Context, buyerID, status string, limit, offset int32) (*orderv1.ListOrdersResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOrderTimeout)
	defer cancel()
	return c.client.ListOrders(ctx, &orderv1.ListOrdersRequest{
		BuyerId: buyerID,
		Status:  status,
		Limit:   limit,
		Offset:  offset,
	})
}

func (c *OrderClient) GetOrder(ctx context.Context, id, requestingUserID string) (*orderv1.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOrderTimeout)
	defer cancel()
	return c.client.GetOrder(ctx, &orderv1.GetOrderRequest{Id: id, RequestingUserId: requestingUserID})
}

func (c *OrderClient) ListEvents(ctx context.Context, orderID string) (*orderv1.ListEventsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOrderTimeout)
	defer cancel()
	return c.client.ListEvents(ctx, &orderv1.ListEventsRequest{OrderId: orderID})
}

func (c *OrderClient) AppendEvent(ctx context.Context, event *orderv1.OrderEvent) (*orderv1.OrderEvent, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOrderTimeout)
	defer cancel()
	return c.client.AppendEvent(ctx, &orderv1.AppendEventRequest{Event: event})
}

func (c *OrderClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := c.client.HealthCheck(ctx, &emptypb.Empty{})
	return err
}

func (c *OrderClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
