package clients

import (
	"context"
	"fmt"
	"time"

	"api-gateway/internal/config"
	"api-gateway/pkg/logger"

	rfqv1 "github.com/nikitashilov/microblog_grpc/proto/rfq/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/types/known/emptypb"
)

const defaultRFQTimeout = 15 * time.Second

// RFQClient talks to the RFQService hosted by post-service (post = inquiry/RFQ
// during the template repurpose; same gRPC address as PostService).
type RFQClient struct {
	conn   *grpc.ClientConn
	client rfqv1.RFQServiceClient
	logger *logger.Logger
}

func NewRFQClient(addr string, tlsCfg config.GRPCTLSConfig, logger *logger.Logger) (*RFQClient, error) {
	creds, err := buildClientTransportCredentials(tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("build rfq client transport credentials: %w", err)
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
		return nil, fmt.Errorf("connect to rfq gRPC service: %w", err)
	}
	return &RFQClient{
		conn:   conn,
		client: rfqv1.NewRFQServiceClient(conn),
		logger: logger,
	}, nil
}

func (c *RFQClient) CreateRFQ(ctx context.Context, rfq *rfqv1.RFQ) (*rfqv1.RFQ, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultRFQTimeout)
	defer cancel()
	return c.client.CreateRFQ(ctx, &rfqv1.CreateRFQRequest{Rfq: rfq})
}

func (c *RFQClient) GetRFQ(ctx context.Context, id, requestingUserID string) (*rfqv1.RFQ, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultRFQTimeout)
	defer cancel()
	return c.client.GetRFQ(ctx, &rfqv1.GetRFQRequest{Id: id, RequestingUserId: requestingUserID})
}

func (c *RFQClient) ListRFQs(ctx context.Context, buyerID, status string, limit, offset int32) (*rfqv1.ListRFQsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultRFQTimeout)
	defer cancel()
	return c.client.ListRFQs(ctx, &rfqv1.ListRFQsRequest{
		BuyerId: buyerID,
		Status:  status,
		Limit:   limit,
		Offset:  offset,
	})
}

func (c *RFQClient) ListOpenRFQs(ctx context.Context, limit, offset int32) (*rfqv1.ListRFQsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultRFQTimeout)
	defer cancel()
	return c.client.ListOpenRFQs(ctx, &rfqv1.ListOpenRFQsRequest{Limit: limit, Offset: offset})
}

func (c *RFQClient) ListQuotesForRFQ(ctx context.Context, rfqID, requestingUserID string) (*rfqv1.ListQuotesResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultRFQTimeout)
	defer cancel()
	return c.client.ListQuotesForRFQ(ctx, &rfqv1.ListQuotesForRFQRequest{
		RfqId:            rfqID,
		RequestingUserId: requestingUserID,
	})
}

func (c *RFQClient) GetRFQForSupplier(ctx context.Context, rfqID, supplierID string) (*rfqv1.SupplierRFQView, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultRFQTimeout)
	defer cancel()
	return c.client.GetRFQForSupplier(ctx, &rfqv1.GetRFQForSupplierRequest{
		RfqId:      rfqID,
		SupplierId: supplierID,
	})
}

func (c *RFQClient) AddQuote(ctx context.Context, quote *rfqv1.Quote) (*rfqv1.Quote, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultRFQTimeout)
	defer cancel()
	return c.client.AddQuote(ctx, &rfqv1.AddQuoteRequest{Quote: quote})
}

func (c *RFQClient) SubmitManufacturerQuote(ctx context.Context, rfqID, manufacturerID, productID string, priceUSD float64, leadTimeDays int32, validityDate, notes string) (*rfqv1.Quote, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultRFQTimeout)
	defer cancel()
	return c.client.SubmitManufacturerQuote(ctx, &rfqv1.SubmitManufacturerQuoteRequest{
		RfqId:          rfqID,
		ManufacturerId: manufacturerID,
		ProductId:      productID,
		PriceUsd:       priceUSD,
		LeadTimeDays:   leadTimeDays,
		ValidityDate:   validityDate,
		SupplierNotes:  notes,
	})
}

func (c *RFQClient) AcceptQuote(ctx context.Context, rfqID, quoteID, actorID string) (*rfqv1.Quote, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultRFQTimeout)
	defer cancel()
	return c.client.AcceptQuote(ctx, &rfqv1.AcceptQuoteRequest{
		RfqId:   rfqID,
		QuoteId: quoteID,
		ActorId: actorID,
	})
}

func (c *RFQClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := c.client.HealthCheck(ctx, &emptypb.Empty{})
	return err
}

func (c *RFQClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
