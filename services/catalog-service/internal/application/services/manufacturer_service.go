package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	appErrors "catalog-service/internal/application/errors"
	"catalog-service/internal/domain/entities"
	"catalog-service/internal/domain/repositories"
	"catalog-service/internal/infrastructure/postgres"
	"catalog-service/pkg/logger"
)

type ManufacturerService struct {
	repo   repositories.ManufacturerRepository
	logger *logger.Logger
	now    func() time.Time
}

func NewManufacturerService(repo repositories.ManufacturerRepository, log *logger.Logger) *ManufacturerService {
	return &ManufacturerService{repo: repo, logger: log, now: time.Now}
}

func (s *ManufacturerService) CreateManufacturer(ctx context.Context, m *entities.Manufacturer) (*entities.Manufacturer, error) {
	if strings.TrimSpace(m.Name) == "" {
		return nil, appErrors.ErrInvalidRequest
	}
	if strings.TrimSpace(m.UserID) == "" {
		return nil, appErrors.ErrInvalidRequest
	}
	id, err := s.nextID(ctx)
	if err != nil {
		s.logger.Error("catalog: allocate manufacturer id: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	m.ID = id
	m.Status = "active"
	if m.Country == "" {
		m.Country = "CN"
	}
	now := s.now()
	m.CreatedAt = now
	m.UpdatedAt = now

	if err := s.repo.Create(ctx, m); err != nil {
		if errors.Is(err, postgres.ErrManufacturerExists) {
			return nil, appErrors.ErrManufacturerExists
		}
		s.logger.Error("catalog: create manufacturer: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	return m, nil
}

func (s *ManufacturerService) GetManufacturer(ctx context.Context, id string) (*entities.Manufacturer, error) {
	if id == "" {
		return nil, appErrors.ErrInvalidRequest
	}
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, appErrors.ErrManufacturerNotFound
		}
		s.logger.Error("catalog: get manufacturer: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	return m, nil
}

func (s *ManufacturerService) GetManufacturerByUser(ctx context.Context, userID string) (*entities.Manufacturer, error) {
	if userID == "" {
		return nil, appErrors.ErrInvalidRequest
	}
	m, err := s.repo.GetByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, appErrors.ErrManufacturerNotFound
		}
		s.logger.Error("catalog: get manufacturer by user: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	return m, nil
}

func (s *ManufacturerService) UpdateManufacturer(ctx context.Context, m *entities.Manufacturer, actorID string) (*entities.Manufacturer, error) {
	if m.ID == "" || actorID == "" {
		return nil, appErrors.ErrInvalidRequest
	}
	existing, err := s.repo.GetByID(ctx, m.ID)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, appErrors.ErrManufacturerNotFound
		}
		s.logger.Error("catalog: update get manufacturer: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	if existing.UserID != actorID {
		return nil, appErrors.ErrUnauthorizedAccess
	}
	m.UserID = existing.UserID // immutable
	m.Verified = existing.Verified
	m.UpdatedAt = s.now()
	if err := s.repo.Update(ctx, m); err != nil {
		s.logger.Error("catalog: update manufacturer: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	// Reload to get DB-side defaults (timestamps).
	updated, err := s.repo.GetByID(ctx, m.ID)
	if err != nil {
		s.logger.Error("catalog: reload after update: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	return updated, nil
}

func (s *ManufacturerService) ListManufacturers(ctx context.Context, f repositories.ListFilter) ([]*entities.Manufacturer, int32, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	ms, total, err := s.repo.List(ctx, f)
	if err != nil {
		s.logger.Error("catalog: list manufacturers: " + err.Error())
		return nil, 0, appErrors.ErrServiceUnavailable
	}
	return ms, total, nil
}

func (s *ManufacturerService) VerifyManufacturer(ctx context.Context, id string, verified bool) (*entities.Manufacturer, error) {
	// ponytail: admin role gated at gateway; add actor role check here if catalog-service ever gets direct callers
	if id == "" {
		return nil, appErrors.ErrInvalidRequest
	}
	if err := s.repo.SetVerified(ctx, id, verified); err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, appErrors.ErrManufacturerNotFound
		}
		s.logger.Error("catalog: verify manufacturer: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("catalog: reload after verify: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	return m, nil
}

func (s *ManufacturerService) nextID(ctx context.Context) (string, error) {
	seq, err := s.repo.NextSeq(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("MFR-%s-%04d", s.now().UTC().Format("20060102"), seq), nil
}
