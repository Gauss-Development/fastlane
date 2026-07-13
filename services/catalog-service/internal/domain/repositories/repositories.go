package repositories

import (
	"context"

	"catalog-service/internal/domain/entities"
)

// ListFilter holds the optional filters for ListManufacturers.
type ListFilter struct {
	Cluster       string
	ServiceType   string
	AssemblyType  string
	Material      string
	VerifiedOnly  bool
	MinLayersGte  int32 // manufacturer.max_layers >= this
	Limit         int32
	Offset        int32
}

type ManufacturerRepository interface {
	Create(ctx context.Context, m *entities.Manufacturer) error
	GetByID(ctx context.Context, id string) (*entities.Manufacturer, error)
	GetByUser(ctx context.Context, userID string) (*entities.Manufacturer, error)
	Update(ctx context.Context, m *entities.Manufacturer) error
	List(ctx context.Context, f ListFilter) ([]*entities.Manufacturer, int32, error)
	SetVerified(ctx context.Context, id string, verified bool) error
	// NextSeq returns the next sequence value for ID generation.
	NextSeq(ctx context.Context) (int64, error)
}
