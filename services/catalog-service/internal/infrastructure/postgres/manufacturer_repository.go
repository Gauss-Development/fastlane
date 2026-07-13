package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"

	"catalog-service/internal/domain/entities"
	"catalog-service/internal/domain/repositories"
)

// ErrManufacturerExists is returned on unique_violation of user_id.
var ErrManufacturerExists = errors.New("manufacturer profile already exists for this user")

type ManufacturerRepository struct {
	db *sql.DB
}

func NewManufacturerRepository(db *sql.DB) *ManufacturerRepository {
	return &ManufacturerRepository{db: db}
}

const insertManufacturer = `
INSERT INTO manufacturers (
    id, user_id, name, name_zh, city, country, cluster, description, website,
    service_types, assembly_types, min_layers, max_layers,
    materials, surface_finishes, min_order_qty, max_order_qty,
    lead_time_days, monthly_capacity, smallest_package, certifications,
    verified, verified_at, rating, order_count, on_time_rate,
    contact_email, contact_wechat, status, created_at, updated_at
) VALUES (
    $1,$2,$3,$4,$5,$6,$7,$8,$9,
    $10,$11,$12,$13,
    $14,$15,$16,$17,
    $18,$19,$20,$21,
    $22,$23,$24,$25,$26,
    $27,$28,$29,$30,$31
)`

func (r *ManufacturerRepository) Create(ctx context.Context, m *entities.Manufacturer) error {
	_, err := r.db.ExecContext(ctx, insertManufacturer,
		m.ID, m.UserID, m.Name, nullStr(m.NameZh), nullStr(m.City), m.Country, nullStr(m.Cluster),
		nullStr(m.Description), nullStr(m.Website),
		pq.Array(m.ServiceTypes), pq.Array(m.AssemblyTypes), m.MinLayers, m.MaxLayers,
		pq.Array(m.Materials), pq.Array(m.SurfaceFinishes), m.MinOrderQty, m.MaxOrderQty,
		m.LeadTimeDays, m.MonthlyCapacity, nullStr(m.SmallestPackage), pq.Array(m.Certifications),
		m.Verified, nullTime(m.VerifiedAt), m.Rating, m.OrderCount, m.OnTimeRate,
		nullStr(m.ContactEmail), nullStr(m.ContactWechat), m.Status, m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrManufacturerExists
		}
		return fmt.Errorf("create manufacturer: %w", err)
	}
	return nil
}

func (r *ManufacturerRepository) GetByID(ctx context.Context, id string) (*entities.Manufacturer, error) {
	const q = `SELECT ` + manufacturerCols + ` FROM manufacturers WHERE id = $1`
	row := r.db.QueryRowContext(ctx, q, id)
	return scanManufacturer(row)
}

func (r *ManufacturerRepository) GetByUser(ctx context.Context, userID string) (*entities.Manufacturer, error) {
	const q = `SELECT ` + manufacturerCols + ` FROM manufacturers WHERE user_id = $1`
	row := r.db.QueryRowContext(ctx, q, userID)
	return scanManufacturer(row)
}

func (r *ManufacturerRepository) Update(ctx context.Context, m *entities.Manufacturer) error {
	const q = `
UPDATE manufacturers SET
    name=$2, name_zh=$3, city=$4, country=$5, cluster=$6, description=$7, website=$8,
    service_types=$9, assembly_types=$10, min_layers=$11, max_layers=$12,
    materials=$13, surface_finishes=$14, min_order_qty=$15, max_order_qty=$16,
    lead_time_days=$17, monthly_capacity=$18, smallest_package=$19, certifications=$20,
    rating=$21, order_count=$22, on_time_rate=$23,
    contact_email=$24, contact_wechat=$25, status=$26, updated_at=now()
WHERE id=$1`
	res, err := r.db.ExecContext(ctx, q,
		m.ID,
		m.Name, nullStr(m.NameZh), nullStr(m.City), m.Country, nullStr(m.Cluster),
		nullStr(m.Description), nullStr(m.Website),
		pq.Array(m.ServiceTypes), pq.Array(m.AssemblyTypes), m.MinLayers, m.MaxLayers,
		pq.Array(m.Materials), pq.Array(m.SurfaceFinishes), m.MinOrderQty, m.MaxOrderQty,
		m.LeadTimeDays, m.MonthlyCapacity, nullStr(m.SmallestPackage), pq.Array(m.Certifications),
		m.Rating, m.OrderCount, m.OnTimeRate,
		nullStr(m.ContactEmail), nullStr(m.ContactWechat), m.Status,
	)
	if err != nil {
		return fmt.Errorf("update manufacturer: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update manufacturer rows affected: %w", err)
	}
	if n == 0 {
		return ErrNoRows
	}
	return nil
}

func (r *ManufacturerRepository) List(ctx context.Context, f repositories.ListFilter) ([]*entities.Manufacturer, int32, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	n := 1

	add := func(clause string, v interface{}) {
		where += fmt.Sprintf(" AND %s", fmt.Sprintf(clause, fmt.Sprintf("$%d", n)))
		args = append(args, v)
		n++
	}

	if f.Cluster != "" {
		add("cluster = %s", f.Cluster)
	}
	if f.VerifiedOnly {
		where += " AND verified = true"
	}
	if f.ServiceType != "" {
		add("%s = ANY(service_types)", f.ServiceType)
	}
	if f.AssemblyType != "" {
		add("%s = ANY(assembly_types)", f.AssemblyType)
	}
	if f.Material != "" {
		add("%s = ANY(materials)", f.Material)
	}
	if f.MinLayersGte > 0 {
		add("max_layers >= %s", f.MinLayersGte)
	}

	countQ := "SELECT COUNT(*) FROM manufacturers " + where
	var total int32
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count manufacturers: %w", err)
	}

	listQ := fmt.Sprintf(
		"SELECT %s FROM manufacturers %s ORDER BY verified DESC, rating DESC LIMIT $%d OFFSET $%d",
		manufacturerCols, where, n, n+1,
	)
	args = append(args, f.Limit, f.Offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list manufacturers: %w", err)
	}
	defer rows.Close()

	var out []*entities.Manufacturer
	for rows.Next() {
		m, err := scanManufacturerRow(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan manufacturer: %w", err)
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate manufacturers: %w", err)
	}
	return out, total, nil
}

func (r *ManufacturerRepository) SetVerified(ctx context.Context, id string, verified bool) error {
	var q string
	var args []interface{}
	if verified {
		q = `UPDATE manufacturers SET verified=true, verified_at=now(), updated_at=now() WHERE id=$1`
		args = []interface{}{id}
	} else {
		q = `UPDATE manufacturers SET verified=false, verified_at=NULL, updated_at=now() WHERE id=$1`
		args = []interface{}{id}
	}
	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("set verified: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("set verified rows affected: %w", err)
	}
	if n == 0 {
		return ErrNoRows
	}
	return nil
}

func (r *ManufacturerRepository) NextSeq(ctx context.Context) (int64, error) {
	var seq int64
	if err := r.db.QueryRowContext(ctx, "SELECT nextval('manufacturer_id_seq')").Scan(&seq); err != nil {
		return 0, fmt.Errorf("next manufacturer seq: %w", err)
	}
	return seq, nil
}

// manufacturerCols is the canonical SELECT column list, matching scanManufacturer order.
const manufacturerCols = `
    id, user_id, name, COALESCE(name_zh,''), COALESCE(city,''), country, COALESCE(cluster,''),
    COALESCE(description,''), COALESCE(website,''),
    service_types, assembly_types, min_layers, max_layers,
    materials, surface_finishes, min_order_qty, max_order_qty,
    lead_time_days, monthly_capacity, COALESCE(smallest_package,''), certifications,
    verified, verified_at, rating, order_count, on_time_rate,
    COALESCE(contact_email,''), COALESCE(contact_wechat,''), status, created_at, updated_at`

// scannerRow is satisfied by both *sql.Row and *sql.Rows.
type scannerRow interface {
	Scan(dest ...interface{}) error
}

func scanManufacturer(row *sql.Row) (*entities.Manufacturer, error) {
	m, err := scanManufacturerRow(row)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNoRows
		}
		return nil, fmt.Errorf("get manufacturer: %w", err)
	}
	return m, nil
}

func scanManufacturerRow(row scannerRow) (*entities.Manufacturer, error) {
	m := &entities.Manufacturer{}
	var verifiedAt sql.NullTime
	err := row.Scan(
		&m.ID, &m.UserID, &m.Name, &m.NameZh, &m.City, &m.Country, &m.Cluster,
		&m.Description, &m.Website,
		pq.Array(&m.ServiceTypes), pq.Array(&m.AssemblyTypes), &m.MinLayers, &m.MaxLayers,
		pq.Array(&m.Materials), pq.Array(&m.SurfaceFinishes), &m.MinOrderQty, &m.MaxOrderQty,
		&m.LeadTimeDays, &m.MonthlyCapacity, &m.SmallestPackage, pq.Array(&m.Certifications),
		&m.Verified, &verifiedAt, &m.Rating, &m.OrderCount, &m.OnTimeRate,
		&m.ContactEmail, &m.ContactWechat, &m.Status, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if verifiedAt.Valid {
		t := verifiedAt.Time
		m.VerifiedAt = &t
	}
	return m, nil
}

// isUniqueViolation returns true for Postgres error code 23505 (unique_violation).
func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}
