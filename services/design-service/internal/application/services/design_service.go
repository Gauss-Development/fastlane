package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	appErrors "design-service/internal/application/errors"
	"design-service/internal/domain/entities"
	"design-service/internal/domain/repositories"
	"design-service/internal/infrastructure/postgres"
	"design-service/internal/infrastructure/storage"
	"design-service/pkg/logger"
)

type DesignService struct {
	projects repositories.ProjectRepository
	files    repositories.FileRepository
	ndas     repositories.NDARepository
	storage  *storage.Client
	logger   *logger.Logger
	now      func() time.Time
}

func NewDesignService(
	projects repositories.ProjectRepository,
	files repositories.FileRepository,
	ndas repositories.NDARepository,
	storage *storage.Client,
	logger *logger.Logger,
) *DesignService {
	return &DesignService{
		projects: projects,
		files:    files,
		ndas:     ndas,
		storage:  storage,
		logger:   logger,
		now:      time.Now,
	}
}

// CreateProject persists a new project; id/status/timestamps assigned server-side.
func (s *DesignService) CreateProject(ctx context.Context, p *entities.Project) (*entities.Project, error) {
	if strings.TrimSpace(p.OwnerID) == "" || strings.TrimSpace(p.Title) == "" {
		return nil, appErrors.ErrInvalidRequest
	}
	id, err := s.nextID(ctx, "PRJ", s.projects.NextProjectSeq)
	if err != nil {
		s.logger.Error("design: allocate project id: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	p.ID = id
	p.Status = "draft"
	if p.Category == "" {
		p.Category = "pcba"
	}
	if err := s.projects.Create(ctx, p); err != nil {
		s.logger.Error("design: create project: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	return p, nil
}

// GetProject enforces owner-or-NDA'd-manufacturer access.
func (s *DesignService) GetProject(ctx context.Context, id, actorID string) (*entities.Project, error) {
	if id == "" || actorID == "" {
		return nil, appErrors.ErrInvalidRequest
	}
	p, err := s.projects.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, appErrors.ErrProjectNotFound
		}
		s.logger.Error("design: get project: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	if p.OwnerID != actorID {
		ok, ndaErr := s.ndas.HasAccepted(ctx, id, actorID)
		if ndaErr != nil {
			s.logger.Error("design: check nda: " + ndaErr.Error())
			return nil, appErrors.ErrServiceUnavailable
		}
		if !ok {
			return nil, appErrors.ErrUnauthorizedAccess
		}
	}
	return p, nil
}

// ListProjects returns projects owned by ownerID, filtered by optional status.
func (s *DesignService) ListProjects(ctx context.Context, ownerID, status string, limit, offset int32) ([]*entities.Project, int32, error) {
	if ownerID == "" {
		return nil, 0, appErrors.ErrInvalidRequest
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	ps, total, err := s.projects.ListByOwner(ctx, ownerID, status, int(limit), int(offset))
	if err != nil {
		s.logger.Error("design: list projects: " + err.Error())
		return nil, 0, appErrors.ErrServiceUnavailable
	}
	return ps, total, nil
}

// RequestUploadURL creates a pending DesignFile row and returns a presigned PUT URL.
// Only the project owner may upload.
func (s *DesignService) RequestUploadURL(ctx context.Context, projectID, actorID, kind, filename, contentType string) (*entities.DesignFile, string, int, error) {
	if projectID == "" || actorID == "" || filename == "" || kind == "" {
		return nil, "", 0, appErrors.ErrInvalidRequest
	}
	p, err := s.projects.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, "", 0, appErrors.ErrProjectNotFound
		}
		s.logger.Error("design: request upload get project: " + err.Error())
		return nil, "", 0, appErrors.ErrServiceUnavailable
	}
	if p.OwnerID != actorID {
		return nil, "", 0, appErrors.ErrUnauthorizedAccess
	}

	fileID, err := s.nextID(ctx, "FILE", s.files.NextFileSeq)
	if err != nil {
		s.logger.Error("design: allocate file id: " + err.Error())
		return nil, "", 0, appErrors.ErrServiceUnavailable
	}

	objectKey := fmt.Sprintf("projects/%s/%s/%s", projectID, fileID, filename)
	f := &entities.DesignFile{
		ID:          fileID,
		ProjectID:   projectID,
		Kind:        kind,
		Filename:    filename,
		Version:     1,
		ObjectKey:   objectKey,
		ContentType: contentType,
		UploadedBy:  actorID,
		Status:      "pending",
	}
	if err := s.files.Create(ctx, f); err != nil {
		s.logger.Error("design: create file row: " + err.Error())
		return nil, "", 0, appErrors.ErrServiceUnavailable
	}

	uploadURL, expiresIn, err := s.storage.PresignPut(ctx, objectKey, contentType)
	if err != nil {
		s.logger.Error("design: presign put: " + err.Error())
		return nil, "", 0, appErrors.ErrServiceUnavailable
	}
	return f, uploadURL, expiresIn, nil
}

// ConfirmUpload transitions the file from pending to committed. Owner-only.
func (s *DesignService) ConfirmUpload(ctx context.Context, fileID, actorID, contentSHA256 string, sizeBytes int64) (*entities.DesignFile, error) {
	if fileID == "" || actorID == "" {
		return nil, appErrors.ErrInvalidRequest
	}
	f, err := s.files.GetByID(ctx, fileID)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, appErrors.ErrFileNotFound
		}
		s.logger.Error("design: confirm upload get file: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	p, err := s.projects.GetByID(ctx, f.ProjectID)
	if err != nil {
		s.logger.Error("design: confirm upload get project: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	if p.OwnerID != actorID {
		return nil, appErrors.ErrUnauthorizedAccess
	}

	committed, err := s.files.ConfirmUpload(ctx, fileID, contentSHA256, sizeBytes)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, appErrors.ErrFileNotFound
		}
		s.logger.Error("design: confirm upload update: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	return committed, nil
}

// ListFiles enforces owner-or-NDA'd-manufacturer access.
func (s *DesignService) ListFiles(ctx context.Context, projectID, actorID string) ([]*entities.DesignFile, error) {
	if projectID == "" || actorID == "" {
		return nil, appErrors.ErrInvalidRequest
	}
	if _, err := s.GetProject(ctx, projectID, actorID); err != nil {
		return nil, err
	}
	fs, err := s.files.ListByProject(ctx, projectID)
	if err != nil {
		s.logger.Error("design: list files: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	return fs, nil
}

// RequestDownloadURL issues a presigned GET. Owner or NDA-accepted manufacturer.
func (s *DesignService) RequestDownloadURL(ctx context.Context, fileID, actorID string) (string, string, int, error) {
	if fileID == "" || actorID == "" {
		return "", "", 0, appErrors.ErrInvalidRequest
	}
	f, err := s.files.GetByID(ctx, fileID)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return "", "", 0, appErrors.ErrFileNotFound
		}
		s.logger.Error("design: download get file: " + err.Error())
		return "", "", 0, appErrors.ErrServiceUnavailable
	}
	// Reuse GetProject for unified authz.
	if _, err := s.GetProject(ctx, f.ProjectID, actorID); err != nil {
		return "", "", 0, err
	}

	downloadURL, expiresIn, err := s.storage.PresignGet(ctx, f.ObjectKey, f.Filename)
	if err != nil {
		s.logger.Error("design: presign get: " + err.Error())
		return "", "", 0, appErrors.ErrServiceUnavailable
	}
	return downloadURL, f.Filename, expiresIn, nil
}

// InviteManufacturer creates a pending NDA row inviting the manufacturer to the
// project. Owner-only; the invite is a prerequisite for AcceptNDA.
func (s *DesignService) InviteManufacturer(ctx context.Context, projectID, manufacturerID, actorID string) (*entities.NDA, error) {
	if projectID == "" || manufacturerID == "" || actorID == "" {
		return nil, appErrors.ErrInvalidRequest
	}
	p, err := s.projects.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, appErrors.ErrProjectNotFound
		}
		s.logger.Error("design: invite get project: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	if p.OwnerID != actorID {
		return nil, appErrors.ErrUnauthorizedAccess
	}
	out, err := s.ndas.CreateInvite(ctx, projectID, manufacturerID)
	if err != nil {
		s.logger.Error("design: create invite: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	return out, nil
}

// AcceptNDA moves an existing (invited) NDA row to accepted. A manufacturer must
// have been invited first — no invite means no access.
func (s *DesignService) AcceptNDA(ctx context.Context, projectID, manufacturerID, ndaVersion, acceptedIP string) (*entities.NDA, error) {
	if projectID == "" || manufacturerID == "" {
		return nil, appErrors.ErrInvalidRequest
	}
	// Must have a prior invite; absence is treated as not authorized (not a 404,
	// so a random manufacturer can't probe which project ids exist).
	existing, err := s.ndas.GetByProjectAndManufacturer(ctx, projectID, manufacturerID)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, appErrors.ErrUnauthorizedAccess
		}
		s.logger.Error("design: accept nda get: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}

	existing.Status = "accepted"
	existing.NDAVersion = ndaVersion
	existing.AcceptedIP = acceptedIP
	out, err := s.ndas.Upsert(ctx, existing)
	if err != nil {
		s.logger.Error("design: accept nda upsert: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	return out, nil
}

// GetNDAStatus returns the NDA for the (project, manufacturer) pair, or a
// pending/absent representation when no row exists.
func (s *DesignService) GetNDAStatus(ctx context.Context, projectID, manufacturerID string) (*entities.NDA, error) {
	if projectID == "" || manufacturerID == "" {
		return nil, appErrors.ErrInvalidRequest
	}
	nda, err := s.ndas.GetByProjectAndManufacturer(ctx, projectID, manufacturerID)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, appErrors.ErrNDANotFound
		}
		s.logger.Error("design: get nda: " + err.Error())
		return nil, appErrors.ErrServiceUnavailable
	}
	return nda, nil
}

// nextID formats PRJ-YYYYMMDD-NNNN / FILE-YYYYMMDD-NNNN. NNNN grows past 4
// digits rather than wrapping (mirrors rfq_service.nextID).
func (s *DesignService) nextID(ctx context.Context, prefix string, next func(context.Context) (int64, error)) (string, error) {
	seq, err := next(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s-%04d", prefix, s.now().UTC().Format("20060102"), seq), nil
}
