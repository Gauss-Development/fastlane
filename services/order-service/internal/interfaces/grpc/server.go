package grpc

import (
	"context"
	"fmt"
	"time"

	appErrors "order-service/internal/application/errors"
	"order-service/internal/application/services"
	"order-service/internal/domain/entities"
	"order-service/pkg/logger"

	orderv1 "github.com/nikitashilov/microblog_grpc/proto/order/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// OrderServer implements orderv1.OrderServiceServer.
type OrderServer struct {
	orderv1.UnimplementedOrderServiceServer
	svc    *services.OrderService
	logger *logger.Logger
}

func NewOrderServer(svc *services.OrderService, log *logger.Logger) *OrderServer {
	return &OrderServer{svc: svc, logger: log}
}

func (s *OrderServer) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.Order, error) {
	p := req.GetOrder()
	if p == nil {
		return nil, status.Error(codes.InvalidArgument, "order is required")
	}
	// CreateOrder via proto is rarely used (consumer is the primary creator).
	// Build a minimal QuoteAcceptedEvent from the proto fields.
	evt := services.QuoteAcceptedEvent{
		RFQID:      p.GetRfqId(),
		QuoteID:    p.GetQuoteId(),
		BuyerID:    p.GetBuyerId(),
		SupplierID: p.GetSupplierId(),
		PriceUSD:   p.GetTotalUsd(),
		Qty:        1,
		AcceptedAt: time.Now(),
	}
	o, err := s.svc.CreateOrderFromQuote(ctx, evt)
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return toProto(o), nil
}

func (s *OrderServer) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.Order, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.GetRequestingUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "requesting_user_id is required")
	}
	o, err := s.svc.GetOrder(ctx, req.GetId(), req.GetRequestingUserId())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return toProto(o), nil
}

func (s *OrderServer) ListOrders(ctx context.Context, req *orderv1.ListOrdersRequest) (*orderv1.ListOrdersResponse, error) {
	os, total, err := s.svc.ListOrders(ctx, req.GetBuyerId(), req.GetStatus(), req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	out := make([]*orderv1.Order, 0, len(os))
	for _, o := range os {
		out = append(out, toProto(o))
	}
	return &orderv1.ListOrdersResponse{Orders: out, Total: total}, nil
}

func (s *OrderServer) AppendEvent(ctx context.Context, req *orderv1.AppendEventRequest) (*orderv1.OrderEvent, error) {
	p := req.GetEvent()
	if p == nil {
		return nil, status.Error(codes.InvalidArgument, "event is required")
	}
	e := eventFromProto(p)
	inserted, err := s.svc.AppendEvent(ctx, e)
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	return eventToProto(inserted), nil
}

func (s *OrderServer) ListEvents(ctx context.Context, req *orderv1.ListEventsRequest) (*orderv1.ListEventsResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	evts, err := s.svc.ListEvents(ctx, req.GetOrderId())
	if err != nil {
		return nil, s.toGRPCError(err)
	}
	out := make([]*orderv1.OrderEvent, 0, len(evts))
	for _, e := range evts {
		out = append(out, eventToProto(e))
	}
	return &orderv1.ListEventsResponse{Events: out}, nil
}

func (s *OrderServer) HealthCheck(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *OrderServer) toGRPCError(err error) error {
	if err == nil {
		return nil
	}
	if oe, ok := err.(*appErrors.OrderError); ok {
		switch oe.Code {
		case appErrors.ErrOrderNotFound.Code:
			return status.Error(codes.NotFound, oe.Message)
		case appErrors.ErrOrderExists.Code:
			return status.Error(codes.AlreadyExists, oe.Message)
		case appErrors.ErrUnauthorizedAccess.Code:
			return status.Error(codes.PermissionDenied, oe.Message)
		case appErrors.ErrInvalidRequest.Code:
			return status.Error(codes.InvalidArgument, oe.Message)
		case appErrors.ErrInvalidTransition.Code:
			return status.Error(codes.FailedPrecondition, oe.Message)
		default:
			return status.Error(codes.Internal, oe.Message)
		}
	}
	s.logger.Error(fmt.Sprintf("order grpc: unexpected error: %v", err))
	return status.Error(codes.Internal, "internal server error")
}

func toProto(o *entities.Order) *orderv1.Order {
	if o == nil {
		return nil
	}
	p := &orderv1.Order{
		Id:                 o.ID,
		BuyerId:            o.BuyerID,
		SupplierId:         o.SupplierID,
		QuoteId:            o.QuoteID,
		RfqId:              o.RFQID,
		Status:             o.Status,
		PaymentStatus:      o.PaymentStatus,
		QcStatus:           o.QCStatus,
		TotalUsd:           o.TotalUSD,
		ShippingAddress:    o.ShippingAddress,
		ShippingCity:       o.ShippingCity,
		ShippingCountry:    o.ShippingCountry,
		WarrantyUntil:      o.WarrantyUntil,
		CancellationReason: o.CancellationReason,
		CreatedAt:          toTS(o.CreatedAt),
		UpdatedAt:          toTS(o.UpdatedAt),
	}
	if o.CancelledAt != nil {
		p.CancelledAt = toTS(*o.CancelledAt)
	}
	return p
}

func eventToProto(e *entities.OrderEvent) *orderv1.OrderEvent {
	if e == nil {
		return nil
	}
	p := &orderv1.OrderEvent{
		Id:         e.ID,
		OrderId:    e.OrderID,
		EventType:  e.EventType,
		FromStatus: e.FromStatus,
		ToStatus:   e.ToStatus,
		ActorId:    e.ActorID,
		ActorType:  e.ActorType,
		OccurredTz: e.OccurredTZ,
		Location:   e.Location,
		Notes:      e.Notes,
		// ponytail: payload/documents passed as empty struct; map jsonb→structpb when UI needs it
		Payload:   &structpb.Struct{},
		Documents: &structpb.Struct{},
	}
	if !e.OccurredAt.IsZero() {
		p.OccurredAt = toTS(e.OccurredAt)
	}
	if !e.CreatedAt.IsZero() {
		p.CreatedAt = toTS(e.CreatedAt)
	}
	return p
}

func eventFromProto(p *orderv1.OrderEvent) *entities.OrderEvent {
	e := &entities.OrderEvent{
		OrderID:    p.GetOrderId(),
		EventType:  p.GetEventType(),
		FromStatus: p.GetFromStatus(),
		ToStatus:   p.GetToStatus(),
		ActorID:    p.GetActorId(),
		ActorType:  p.GetActorType(),
		OccurredTZ: p.GetOccurredTz(),
		Location:   p.GetLocation(),
		Notes:      p.GetNotes(),
	}
	if p.GetOccurredAt() != nil {
		e.OccurredAt = p.GetOccurredAt().AsTime()
	}
	if e.OccurredTZ == "" {
		e.OccurredTZ = "UTC"
	}
	return e
}

func toTS(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
