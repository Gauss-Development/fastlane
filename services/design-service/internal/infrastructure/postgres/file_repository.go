package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"design-service/internal/domain/entities"
)

type FileRepository struct {
	db *sql.DB
}

func NewFileRepository(db *sql.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, f *entities.DesignFile) error {
	const q = `
		INSERT INTO design_files
			(id, project_id, kind, filename, version, object_key, size_bytes, content_type, uploaded_by, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, q,
		f.ID, f.ProjectID, f.Kind, f.Filename, f.Version, f.ObjectKey,
		f.SizeBytes, nullIfEmpty(f.ContentType), f.UploadedBy, f.Status, now)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	f.CreatedAt = now
	return nil
}

func (r *FileRepository) GetByID(ctx context.Context, id string) (*entities.DesignFile, error) {
	const q = `
		SELECT id, project_id, kind, filename, version,
		       COALESCE(content_sha256,''), object_key, COALESCE(size_bytes,0),
		       COALESCE(content_type,''), uploaded_by, status, created_at
		FROM design_files WHERE id = $1`
	f := &entities.DesignFile{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&f.ID, &f.ProjectID, &f.Kind, &f.Filename, &f.Version,
		&f.ContentSHA256, &f.ObjectKey, &f.SizeBytes,
		&f.ContentType, &f.UploadedBy, &f.Status, &f.CreatedAt)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNoRows
		}
		return nil, fmt.Errorf("get file: %w", err)
	}
	return f, nil
}

func (r *FileRepository) ListByProject(ctx context.Context, projectID string) ([]*entities.DesignFile, error) {
	const q = `
		SELECT id, project_id, kind, filename, version,
		       COALESCE(content_sha256,''), object_key, COALESCE(size_bytes,0),
		       COALESCE(content_type,''), uploaded_by, status, created_at
		FROM design_files WHERE project_id = $1 ORDER BY created_at`
	rows, err := r.db.QueryContext(ctx, q, projectID)
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}
	defer rows.Close()

	var out []*entities.DesignFile
	for rows.Next() {
		f := &entities.DesignFile{}
		if err := rows.Scan(
			&f.ID, &f.ProjectID, &f.Kind, &f.Filename, &f.Version,
			&f.ContentSHA256, &f.ObjectKey, &f.SizeBytes,
			&f.ContentType, &f.UploadedBy, &f.Status, &f.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan file: %w", err)
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *FileRepository) ConfirmUpload(ctx context.Context, fileID, contentSHA256 string, sizeBytes int64) (*entities.DesignFile, error) {
	const q = `
		UPDATE design_files
		SET status = 'committed', content_sha256 = $2, size_bytes = $3
		WHERE id = $1 AND status = 'pending'
		RETURNING id, project_id, kind, filename, version,
		          COALESCE(content_sha256,''), object_key, COALESCE(size_bytes,0),
		          COALESCE(content_type,''), uploaded_by, status, created_at`
	f := &entities.DesignFile{}
	err := r.db.QueryRowContext(ctx, q, fileID, nullIfEmpty(contentSHA256), sizeBytes).Scan(
		&f.ID, &f.ProjectID, &f.Kind, &f.Filename, &f.Version,
		&f.ContentSHA256, &f.ObjectKey, &f.SizeBytes,
		&f.ContentType, &f.UploadedBy, &f.Status, &f.CreatedAt)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNoRows
		}
		return nil, fmt.Errorf("confirm upload: %w", err)
	}
	return f, nil
}

func (r *FileRepository) NextFileSeq(ctx context.Context) (int64, error) {
	var seq int64
	if err := r.db.QueryRowContext(ctx, "SELECT nextval('file_id_seq')").Scan(&seq); err != nil {
		return 0, fmt.Errorf("next file seq: %w", err)
	}
	return seq, nil
}
