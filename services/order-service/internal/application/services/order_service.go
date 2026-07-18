package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	appErrors "order-service/internal/application/errors"
	"order-service/internal/domain/entities"
	"order-service/internal/domain/repositories"
	"order-service/internal/infrastructure/postgres"
	"order-service/pkg/logger"
)

// QuoteAcceptedEvent mirrors the JSON payload from post-service's quote.accepted.
// json tags MUST match the producer exactly.
type QuoteAcceptedEvent struct {
	RFQID           string    `json:"rfq_id"`
	QuoteID         string    `json:"quote_id"`
	BuyerID         string    `json:"buyer_id"`
	BuyerEmail      string    `json:"buyer_email"`
	BuyerCompany    string    `json:"buyer_company"`
	QueryText       string    `json:"query_text"`
	SupplierID      string    `json:"supplier_id"`
	ManufacturerID  string    `json:"manufacturer_id"`
	ProductID       string    `json:"product_id"`
	PriceUSD        float64   `json:"price_usd"`
	Qty             int32     `json:"qty"`
	ShippingAddress string    `json:"shipping_address"`
	AcceptedAt      time.Time `json:"accepted_at"`
}

type OrderService struct {
	orders repositories.OrderRepository
	events repositories.OrderEventRepository
	logger *logger.Logger
	now    func() time.Time
}

func NewOrderService(orders repositories.OrderRepository, events repositories.OrderEventRepository, log *logger.Logger) *OrderService {
	return &OrderService{orders: orders, events: events, logger: log, now: time.Now}
}

// CreateOrderFromQuote is idempotent: if an order already exists for the quote it is returned unchanged.
func (s *OrderService) CreateOrderFromQuote(ctx context.Context, evt QuoteAcceptedEvent) (*entities.Order, error) {
	existing, err := s.orders.GetByQuoteID(ctx, evt.QuoteID)
	if err != nil && !errors.Is(err, postgres.ErrNoRows) {
		s.logger.Error("order: lookup by quote_id: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	if existing != nil {
		s.logger.Info("order: CreateOrderFromQuote skipped, already exists: " + existing.ID)
		return existing, nil
	}

	id, err := s.nextID(ctx)
	if err != nil {
		s.logger.Error("order: allocate id: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}

	qty := evt.Qty
	if qty <= 0 {
		qty = 1
	}
	supplierID := evt.SupplierID
	if supplierID == "" {
		supplierID = evt.ManufacturerID
	}

	now := s.now()
	o := &entities.Order{
		ID:              id,
		BuyerID:         evt.BuyerID,
		SupplierID:      supplierID,
		QuoteID:         evt.QuoteID,
		RFQID:           evt.RFQID,
		Status:          "pending_payment",
		PaymentStatus:   "unpaid",
		TotalUSD:        evt.PriceUSD * float64(qty),
		ShippingAddress: evt.ShippingAddress,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.orders.Create(ctx, o); err != nil {
		s.logger.Error("order: create: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}

	// Insert the initial order_created event.
	_, err = s.events.Insert(ctx, &entities.OrderEvent{
		OrderID:    id,
		EventType:  "order_created",
		FromStatus: "",
		ToStatus:   "pending_payment",
		ActorID:    evt.BuyerID,
		ActorType:  "system",
		OccurredAt: evt.AcceptedAt,
		OccurredTZ: "UTC",
		Notes:      "Order created from accepted quote " + evt.QuoteID,
	})
	if err != nil {
		// Non-fatal: the order exists; the event is idempotent via unique constraint.
		s.logger.Warn("order: insert order_created event: " + err.Error())
	}

	return o, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id, requestingUserID string) (*entities.Order, error) {
	if id == "" || requestingUserID == "" {
		return nil, appErrors.ErrInvalidRequest
	}
	o, err := s.orders.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, appErrors.ErrOrderNotFound
		}
		s.logger.Error("order: get: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	if o.BuyerID != requestingUserID {
		return nil, appErrors.ErrUnauthorizedAccess
	}
	return o, nil
}

func (s *OrderService) ListOrders(ctx context.Context, buyerID, status string, limit, offset int32) ([]*entities.Order, int32, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	os, total, err := s.orders.List(ctx, repositories.ListOrdersFilter{
		BuyerID: buyerID,
		Status:  status,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		s.logger.Error("order: list: " + err.Error())
		return nil, 0, appErrors.ErrServiceUnavailable
	}
	return os, total, nil
}

func (s *OrderService) AppendEvent(ctx context.Context, e *entities.OrderEvent) (*entities.OrderEvent, error) {
	if e.OrderID == "" || e.EventType == "" || e.ActorType == "" {
		return nil, appErrors.ErrInvalidRequest
	}
	if e.OccurredTZ == "" {
		e.OccurredTZ = "UTC"
	}
	if e.OccurredAt.IsZero() {
		e.OccurredAt = s.now()
	}

	// If to_status is set, validate the transition before writing.
	if e.ToStatus != "" {
		o, err := s.orders.GetByID(ctx, e.OrderID)
		if err != nil {
			if errors.Is(err, postgres.ErrNoRows) {
				return nil, appErrors.ErrOrderNotFound
			}
			s.logger.Error("order: append event get order: " + err.Error())
			return nil, appErrors.ErrServiceUnavailable
		}
		if !entities.CanTransition(o.Status, e.ToStatus) {
			return nil, appErrors.ErrInvalidTransition
		}
		// Write event first; then update status.
		inserted, err := s.events.Insert(ctx, e)
		if err != nil {
			s.logger.Error("order: insert event: " + err.Error())
			return nil, appErrors.ErrServiceUnavailable
		}
		if err := s.orders.UpdateStatus(ctx, e.OrderID, e.ToStatus, o.PaymentStatus, o.QCStatus); err != nil {
			s.logger.Error("order: update status after event: " + err.Error())
			return nil, appErrors.ErrServiceUnavailable
		}
		return inserted, nil
	}

	inserted, err := s.events.Insert(ctx, e)
	if err != nil {
		s.logger.Error("order: insert event: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	return inserted, nil
}

func (s *OrderService) ListEvents(ctx context.Context, orderID string) ([]*entities.OrderEvent, error) {
	if orderID == "" {
		return nil, appErrors.ErrInvalidRequest
	}
	evts, err := s.events.ListByOrder(ctx, orderID)
	if err != nil {
		s.logger.Error("order: list events: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	return evts, nil
}

func (s *OrderService) nextID(ctx context.Context) (string, error) {
	seq, err := s.orders.NextSeq(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("ORD-%s-%04d-SFO", s.now().UTC().Format("20060102"), seq), nil
}
