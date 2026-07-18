package repositories

import (
	"context"

	"order-service/internal/domain/entities"
)

type ListOrdersFilter struct {
	BuyerID string
	Status  string
	Limit   int32
	Offset  int32
}

type OrderRepository interface {
	Create(ctx context.Context, o *entities.Order) error
	GetByID(ctx context.Context, id string) (*entities.Order, error)
	GetByQuoteID(ctx context.Context, quoteID string) (*entities.Order, error)
	List(ctx context.Context, f ListOrdersFilter) ([]*entities.Order, int32, error)
	UpdateStatus(ctx context.Context, id, status, paymentStatus, qcStatus string) error
	NextSeq(ctx context.Context) (int64, error)
}

type OrderEventRepository interface {
	Insert(ctx context.Context, e *entities.OrderEvent) (*entities.OrderEvent, error)
	ListByOrder(ctx context.Context, orderID string) ([]*entities.OrderEvent, error)
}
