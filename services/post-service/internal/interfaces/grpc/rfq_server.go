package grpc

import (
	"context"
	"encoding/json"
	"time"

	"post-service/internal/application/dto"
	appErrors "post-service/internal/application/errors"
	"post-service/internal/application/services"
	"post-service/internal/domain/entities"
	"post-service/pkg/logger"

	rfqv1 "github.com/nikitashilov/microblog_grpc/proto/rfq/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RFQServer struct {
	rfqv1.UnimplementedRFQServiceServer
	service *services.RFQService
	logger  *logger.Logger
}

func NewRFQServer(service *services.RFQService, logger *logger.Logger) *RFQServer {
	return &RFQServer{service: service, logger: logger}
}

func (s *RFQServer) CreateRFQ(ctx context.Context, req *rfqv1.CreateRFQRequest) (*rfqv1.RFQ, error) {
	rfq := req.GetRfq()
	if rfq == nil || rfq.GetBuyerId() == "" {
		return nil, status.Error(codes.Unauthenticated, appErrors.ErrUnauthorizedAccess.Message)
	}

	var parsedSpecs json.RawMessage
	if rfq.GetParsedSpecs() != nil {
		if raw, err := rfq.GetParsedSpecs().MarshalJSON(); err == nil {
			parsedSpecs = raw
		}
	}

	created, err := s.service.CreateRFQ(ctx, &dto.CreateRFQRequest{
		BuyerID:           rfq.GetBuyerId(),
		BuyerEmail:        rfq.GetBuyerEmail(),
		BuyerCompany:      rfq.GetBuyerCompany(),
		QueryText:         rfq.GetQueryText(),
		ParsedSpecs:       parsedSpecs,
		MatchedProductIDs: rfq.GetMatchedProductIds(),
		Qty:               rfq.GetQty(),
		TargetDate:        rfq.GetTargetDate(),
		ShippingAddress:   rfq.GetShippingAddress(),
		Notes:             rfq.GetNotes(),
		ProjectID:         rfq.GetProjectId(),
	})
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return toProtoRFQ(created), nil
}

func (s *RFQServer) GetRFQ(ctx context.Context, req *rfqv1.GetRFQRequest) (*rfqv1.RFQ, error) {
	if req.GetId() == "" || req.GetRequestingUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, appErrors.ErrInvalidRequest.Message)
	}
	rfq, err := s.service.GetRFQ(ctx, req.GetId(), req.GetRequestingUserId())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return toProtoRFQ(rfq), nil
}

func (s *RFQServer) ListRFQs(ctx context.Context, req *rfqv1.ListRFQsRequest) (*rfqv1.ListRFQsResponse, error) {
	if req.GetBuyerId() == "" {
		return nil, status.Error(codes.InvalidArgument, appErrors.ErrInvalidRequest.Message)
	}
	rfqs, total, err := s.service.ListRFQs(ctx, &dto.ListRFQsRequest{
		BuyerID: req.GetBuyerId(),
		Status:  req.GetStatus(),
		Limit:   req.GetLimit(),
		Offset:  req.GetOffset(),
	})
	if err != nil {
		return nil, s.toGRPCError(err)
	}

	out := make([]*rfqv1.RFQ, 0, len(rfqs))
	for _, rfq := range rfqs {
		out = append(out, toProtoRFQ(rfq))
	}
	return &rfqv1.ListRFQsResponse{Rfqs: out, Total: total}, nil
}

func (s *RFQServer) ListOpenRFQs(ctx context.Context, req *rfqv1.ListOpenRFQsRequest) (*rfqv1.ListRFQsResponse, error) {
	rfqs, total, err := s.service.ListOpenRFQs(ctx, req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	out := make([]*rfqv1.RFQ, 0, len(rfqs))
	for _, rfq := range rfqs {
		out = append(out, toProtoRFQ(rfq))
	}
	return &rfqv1.ListRFQsResponse{Rfqs: out, Total: total}, nil
}

// AddQuote is the supplier quote submission: the gateway resolves the magic
// link token to (rfq_id, supplier_id) and the service updates the pending row.
func (s *RFQServer) AddQuote(ctx context.Context, req *rfqv1.AddQuoteRequest) (*rfqv1.Quote, error) {
	quote := req.GetQuote()
	if quote == nil || quote.GetRfqId() == "" || quote.GetSupplierId() == "" {
		return nil, status.Error(codes.InvalidArgument, appErrors.ErrInvalidRequest.Message)
	}

	submitted, err := s.service.SubmitQuote(ctx, &dto.SubmitQuoteRequest{
		RFQID:         quote.GetRfqId(),
		SupplierID:    quote.GetSupplierId(),
		PriceUSD:      quote.GetPriceUsd(),
		LeadTimeDays:  quote.GetLeadTimeDays(),
		ValidityDate:  quote.GetValidityDate(),
		SupplierNotes: quote.GetSupplierNotes(),
	})
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return toProtoQuote(submitted), nil
}

func (s *RFQServer) ListQuotesForRFQ(ctx context.Context, req *rfqv1.ListQuotesForRFQRequest) (*rfqv1.ListQuotesResponse, error) {
	if req.GetRfqId() == "" || req.GetRequestingUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, appErrors.ErrInvalidRequest.Message)
	}
	quotes, err := s.service.ListQuotesForRFQ(ctx, req.GetRfqId(), req.GetRequestingUserId())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	out := make([]*rfqv1.Quote, 0, len(quotes))
	for _, q := range quotes {
		out = append(out, toProtoQuote(q))
	}
	return &rfqv1.ListQuotesResponse{Quotes: out}, nil
}

func (s *RFQServer) GetRFQForSupplier(ctx context.Context, req *rfqv1.GetRFQForSupplierRequest) (*rfqv1.SupplierRFQView, error) {
	if req.GetRfqId() == "" || req.GetSupplierId() == "" {
		return nil, status.Error(codes.InvalidArgument, appErrors.ErrInvalidRequest.Message)
	}
	rfq, quote, supplierName, err := s.service.GetRFQForSupplier(ctx, req.GetRfqId(), req.GetSupplierId())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return &rfqv1.SupplierRFQView{
		Rfq:          toProtoRFQ(rfq),
		Quote:        toProtoQuote(quote),
		SupplierName: supplierName,
	}, nil
}

func (s *RFQServer) HealthCheck(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *RFQServer) toGRPCError(err error) error {
	if err == nil {
		return nil
	}
	if postErr, ok := err.(*appErrors.PostError); ok {
		switch postErr.Code {
		case appErrors.ErrRFQNotFound.Code, appErrors.ErrQuoteNotFound.Code:
			return status.Error(codes.NotFound, postErr.Message)
		case appErrors.ErrQuoteAlreadyExists.Code:
			return status.Error(codes.AlreadyExists, postErr.Message)
		case appErrors.ErrUnauthorizedAccess.Code:
			return status.Error(codes.PermissionDenied, postErr.Message)
		case appErrors.ErrInvalidRFQData.Code, appErrors.ErrNoMatchedProducts.Code, appErrors.ErrInvalidRequest.Code:
			return status.Error(codes.InvalidArgument, postErr.Message)
		default:
			return status.Error(codes.Internal, postErr.Message)
		}
	}
	s.logger.Error("rfq grpc: unexpected error: " + err.Error())
	return status.Error(codes.Internal, "internal server error")
}

func toProtoRFQ(rfq *entities.RFQ) *rfqv1.RFQ {
	if rfq == nil {
		return nil
	}
	var specs *structpb.Struct
	if len(rfq.ParsedSpecs) > 0 {
		specs = &structpb.Struct{}
		if err := specs.UnmarshalJSON(rfq.ParsedSpecs); err != nil {
			specs = nil
		}
	}
	return &rfqv1.RFQ{
		Id:                rfq.ID,
		BuyerId:           rfq.BuyerID,
		BuyerEmail:        rfq.BuyerEmail,
		BuyerCompany:      rfq.BuyerCompany,
		QueryText:         rfq.QueryText,
		ParsedSpecs:       specs,
		MatchedProductIds: rfq.MatchedProductIDs,
		Status:            rfq.Status,
		Qty:               rfq.Qty,
		TargetDate:        rfq.TargetDate,
		ShippingAddress:   rfq.ShippingAddress,
		Notes:             rfq.Notes,
		ProjectId:         rfq.ProjectID,
		CreatedAt:         toTimestamp(rfq.CreatedAt),
	}
}

func toTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func (s *RFQServer) SubmitManufacturerQuote(ctx context.Context, req *rfqv1.SubmitManufacturerQuoteRequest) (*rfqv1.Quote, error) {
	if req.GetRfqId() == "" || req.GetManufacturerId() == "" {
		return nil, status.Error(codes.InvalidArgument, appErrors.ErrInvalidRequest.Message)
	}
	quote, err := s.service.SubmitManufacturerQuote(ctx, &dto.SubmitManufacturerQuoteRequest{
		RFQID:          req.GetRfqId(),
		ManufacturerID: req.GetManufacturerId(),
		ProductID:      req.GetProductId(),
		PriceUSD:       req.GetPriceUsd(),
		LeadTimeDays:   req.GetLeadTimeDays(),
		ValidityDate:   req.GetValidityDate(),
		SupplierNotes:  req.GetSupplierNotes(),
	})
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return toProtoQuote(quote), nil
}

func (s *RFQServer) AcceptQuote(ctx context.Context, req *rfqv1.AcceptQuoteRequest) (*rfqv1.Quote, error) {
	if req.GetRfqId() == "" || req.GetQuoteId() == "" || req.GetActorId() == "" {
		return nil, status.Error(codes.InvalidArgument, appErrors.ErrInvalidRequest.Message)
	}
	quote, err := s.service.AcceptQuote(ctx, req.GetRfqId(), req.GetQuoteId(), req.GetActorId())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return toProtoQuote(quote), nil
}

func toProtoQuote(quote *entities.Quote) *rfqv1.Quote {
	if quote == nil {
		return nil
	}
	return &rfqv1.Quote{
		Id:             quote.ID,
		RfqId:          quote.RFQID,
		SupplierId:     quote.SupplierID,
		ManufacturerId: quote.ManufacturerID,
		ProductId:      quote.ProductID,
		PriceUsd:       quote.PriceUSD,
		LeadTimeDays:   quote.LeadTimeDays,
		ValidityDate:   quote.ValidityDate,
		SupplierNotes:  quote.SupplierNotes,
		MatchScore:     quote.MatchScore,
		Status:         quote.Status,
		SubmittedAt:    toTimestamp(quote.SubmittedAt),
		CreatedAt:      toTimestamp(quote.CreatedAt),
	}
}
