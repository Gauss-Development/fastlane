package repositories

import (
	"context"

	"design-service/internal/domain/entities"
)

type ProjectRepository interface {
	Create(ctx context.Context, p *entities.Project) error
	GetByID(ctx context.Context, id string) (*entities.Project, error)
	ListByOwner(ctx context.Context, ownerID, status string, limit, offset int) ([]*entities.Project, int32, error)
	// NextProjectSeq returns the next sequence value for ID generation.
	NextProjectSeq(ctx context.Context) (int64, error)
}

type FileRepository interface {
	Create(ctx context.Context, f *entities.DesignFile) error
	GetByID(ctx context.Context, id string) (*entities.DesignFile, error)
	ListByProject(ctx context.Context, projectID string) ([]*entities.DesignFile, error)
	ConfirmUpload(ctx context.Context, fileID, contentSHA256 string, sizeBytes int64) (*entities.DesignFile, error)
	// NextFileSeq returns the next sequence value for ID generation.
	NextFileSeq(ctx context.Context) (int64, error)
}

type NDARepository interface {
	Upsert(ctx context.Context, nda *entities.NDA) (*entities.NDA, error)
	// CreateInvite inserts a pending NDA row (no-op if one already exists) and returns the current row.
	CreateInvite(ctx context.Context, projectID, manufacturerID string) (*entities.NDA, error)
	GetByProjectAndManufacturer(ctx context.Context, projectID, manufacturerID string) (*entities.NDA, error)
	// HasAccepted returns true if the manufacturer has an accepted NDA on the project.
	HasAccepted(ctx context.Context, projectID, manufacturerID string) (bool, error)
}
