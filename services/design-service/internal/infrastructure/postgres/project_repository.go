package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"design-service/internal/domain/entities"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, p *entities.Project) error {
	const q = `
		INSERT INTO projects
			(id, owner_id, title, description, category, status, owner_email, owner_company, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, q,
		p.ID, p.OwnerID, p.Title, nullIfEmpty(p.Description), p.Category, p.Status,
		nullIfEmpty(p.OwnerEmail), nullIfEmpty(p.OwnerCompany), now, now)
	if err != nil {
		return fmt.Errorf("create project: %w", err)
	}
	p.CreatedAt = now
	p.UpdatedAt = now
	return nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*entities.Project, error) {
	const q = `
		SELECT id, owner_id, title, COALESCE(description,''), category, status,
		       COALESCE(owner_email,''), COALESCE(owner_company,''), created_at, updated_at
		FROM projects WHERE id = $1`
	p := &entities.Project{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&p.ID, &p.OwnerID, &p.Title, &p.Description, &p.Category, &p.Status,
		&p.OwnerEmail, &p.OwnerCompany, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNoRows
		}
		return nil, fmt.Errorf("get project: %w", err)
	}
	return p, nil
}

func (r *ProjectRepository) ListByOwner(ctx context.Context, ownerID, status string, limit, offset int) ([]*entities.Project, int32, error) {
	// Build query dynamically to handle optional status filter.
	whereStatus := ""
	args := []interface{}{ownerID}
	if status != "" {
		whereStatus = " AND status = $2"
		args = append(args, status)
	}
	countQ := "SELECT COUNT(*) FROM projects WHERE owner_id = $1" + whereStatus
	listQ := fmt.Sprintf(`
		SELECT id, owner_id, title, COALESCE(description,''), category, status,
		       COALESCE(owner_email,''), COALESCE(owner_company,''), created_at, updated_at
		FROM projects WHERE owner_id = $1%s
		ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		whereStatus, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	var total int32
	if err := r.db.QueryRowContext(ctx, countQ, args[:len(args)-2]...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count projects: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var out []*entities.Project
	for rows.Next() {
		p := &entities.Project{}
		if err := rows.Scan(
			&p.ID, &p.OwnerID, &p.Title, &p.Description, &p.Category, &p.Status,
			&p.OwnerEmail, &p.OwnerCompany, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan project: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate projects: %w", err)
	}
	return out, total, nil
}

func (r *ProjectRepository) NextProjectSeq(ctx context.Context) (int64, error) {
	var seq int64
	if err := r.db.QueryRowContext(ctx, "SELECT nextval('project_id_seq')").Scan(&seq); err != nil {
		return 0, fmt.Errorf("next project seq: %w", err)
	}
	return seq, nil
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
