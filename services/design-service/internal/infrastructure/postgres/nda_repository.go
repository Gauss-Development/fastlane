package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"design-service/internal/domain/entities"
)

type NDARepository struct {
	db *sql.DB
}

func NewNDARepository(db *sql.DB) *NDARepository {
	return &NDARepository{db: db}
}

func (r *NDARepository) Upsert(ctx context.Context, nda *entities.NDA) (*entities.NDA, error) {
	const q = `
		INSERT INTO ndas (id, project_id, manufacturer_id, status, nda_version, accepted_ip, accepted_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (project_id, manufacturer_id) DO UPDATE
		  SET status      = EXCLUDED.status,
		      nda_version = EXCLUDED.nda_version,
		      accepted_ip = EXCLUDED.accepted_ip,
		      accepted_at = EXCLUDED.accepted_at
		RETURNING id, project_id, manufacturer_id, status,
		          COALESCE(nda_version,''), COALESCE(accepted_ip,''), accepted_at, created_at`

	now := time.Now()
	acceptedAt := nda.AcceptedAt
	if nda.Status == "accepted" && acceptedAt == nil {
		acceptedAt = &now
	}

	out := &entities.NDA{}
	var dbAcceptedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, q,
		nda.ID, nda.ProjectID, nda.ManufacturerID, nda.Status,
		nullIfEmpty(nda.NDAVersion), nullIfEmpty(nda.AcceptedIP), acceptedAt, now,
	).Scan(&out.ID, &out.ProjectID, &out.ManufacturerID, &out.Status,
		&out.NDAVersion, &out.AcceptedIP, &dbAcceptedAt, &out.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("upsert nda: %w", err)
	}
	if dbAcceptedAt.Valid {
		t := dbAcceptedAt.Time
		out.AcceptedAt = &t
	}
	return out, nil
}

// CreateInvite inserts a pending NDA row for the (project, manufacturer) pair.
// ON CONFLICT DO NOTHING so an already-accepted invite is never downgraded; the
// current row (pending or accepted) is then returned.
func (r *NDARepository) CreateInvite(ctx context.Context, projectID, manufacturerID string) (*entities.NDA, error) {
	const q = `
		INSERT INTO ndas (id, project_id, manufacturer_id, status, created_at)
		VALUES ($1,$2,$3,'pending',$4)
		ON CONFLICT (project_id, manufacturer_id) DO NOTHING`
	id := fmt.Sprintf("NDA-%s-%s", projectID, manufacturerID)
	if _, err := r.db.ExecContext(ctx, q, id, projectID, manufacturerID, time.Now()); err != nil {
		return nil, fmt.Errorf("create nda invite: %w", err)
	}
	return r.GetByProjectAndManufacturer(ctx, projectID, manufacturerID)
}

func (r *NDARepository) GetByProjectAndManufacturer(ctx context.Context, projectID, manufacturerID string) (*entities.NDA, error) {
	const q = `
		SELECT id, project_id, manufacturer_id, status,
		       COALESCE(nda_version,''), COALESCE(accepted_ip,''), accepted_at, created_at
		FROM ndas WHERE project_id = $1 AND manufacturer_id = $2`
	out := &entities.NDA{}
	var dbAcceptedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, q, projectID, manufacturerID).Scan(
		&out.ID, &out.ProjectID, &out.ManufacturerID, &out.Status,
		&out.NDAVersion, &out.AcceptedIP, &dbAcceptedAt, &out.CreatedAt)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNoRows
		}
		return nil, fmt.Errorf("get nda: %w", err)
	}
	if dbAcceptedAt.Valid {
		t := dbAcceptedAt.Time
		out.AcceptedAt = &t
	}
	return out, nil
}

func (r *NDARepository) HasAccepted(ctx context.Context, projectID, manufacturerID string) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM ndas WHERE project_id=$1 AND manufacturer_id=$2 AND status='accepted')`
	var ok bool
	if err := r.db.QueryRowContext(ctx, q, projectID, manufacturerID).Scan(&ok); err != nil {
		return false, fmt.Errorf("check nda: %w", err)
	}
	return ok, nil
}
