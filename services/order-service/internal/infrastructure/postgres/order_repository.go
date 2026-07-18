package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"order-service/internal/domain/entities"
	"order-service/internal/domain/repositories"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

const orderCols = `
    id, buyer_id, supplier_id, quote_id, rfq_id,
    status, payment_status, COALESCE(qc_status,''),
    total_usd,
    COALESCE(shipping_address,''), COALESCE(shipping_city,''), COALESCE(shipping_country,''),
    COALESCE(warranty_until::text,''),
    cancelled_at, COALESCE(cancellation_reason,''),
    created_at, updated_at`

func (r *OrderRepository) Create(ctx context.Context, o *entities.Order) error {
	const q = `
INSERT INTO orders (
    id, buyer_id, supplier_id, quote_id, rfq_id,
    status, payment_status, qc_status,
    total_usd, shipping_address, shipping_city, shipping_country,
    warranty_until, cancelled_at, cancellation_reason,
    created_at, updated_at
) VALUES (
    $1,$2,$3,$4,$5,
    $6,$7,$8,
    $9,$10,$11,$12,
    $13,$14,$15,
    $16,$17
) ON CONFLICT (quote_id) DO NOTHING`

	var warrantyUntil interface{}
	if o.WarrantyUntil != "" {
		warrantyUntil = o.WarrantyUntil
	}

	_, err := r.db.ExecContext(ctx, q,
		o.ID, o.BuyerID, o.SupplierID, o.QuoteID, o.RFQID,
		o.Status, o.PaymentStatus, nullStr(o.QCStatus),
		o.TotalUSD, nullStr(o.ShippingAddress), nullStr(o.ShippingCity), nullStr(o.ShippingCountry),
		warrantyUntil, nullTimePtr(o.CancelledAt), nullStr(o.CancellationReason),
		o.CreatedAt, o.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create order: %w", err)
	}
	return nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*entities.Order, error) {
	const q = `SELECT ` + orderCols + ` FROM orders WHERE id = $1`
	return scanOrder(r.db.QueryRowContext(ctx, q, id))
}

func (r *OrderRepository) GetByQuoteID(ctx context.Context, quoteID string) (*entities.Order, error) {
	const q = `SELECT ` + orderCols + ` FROM orders WHERE quote_id = $1`
	return scanOrder(r.db.QueryRowContext(ctx, q, quoteID))
}

func (r *OrderRepository) List(ctx context.Context, f repositories.ListOrdersFilter) ([]*entities.Order, int32, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	n := 1

	add := func(clause string, v interface{}) {
		where += fmt.Sprintf(" AND %s", fmt.Sprintf(clause, fmt.Sprintf("$%d", n)))
		args = append(args, v)
		n++
	}

	if f.BuyerID != "" {
		add("buyer_id = %s", f.BuyerID)
	}
	if f.Status != "" {
		add("status = %s", f.Status)
	}

	var total int32
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	listQ := fmt.Sprintf(
		"SELECT %s FROM orders %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		orderCols, where, n, n+1,
	)
	args = append(args, f.Limit, f.Offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var out []*entities.Order
	for rows.Next() {
		o, err := scanOrderRow(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan order: %w", err)
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate orders: %w", err)
	}
	return out, total, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id, status, paymentStatus, qcStatus string) error {
	const q = `
UPDATE orders SET
    status = $2,
    payment_status = $3,
    qc_status = CASE WHEN $4 = '' THEN qc_status ELSE $4::text END,
    updated_at = now()
WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, id, status, paymentStatus, qcStatus)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update order rows affected: %w", err)
	}
	if n == 0 {
		return ErrNoRows
	}
	return nil
}

func (r *OrderRepository) NextSeq(ctx context.Context) (int64, error) {
	var seq int64
	if err := r.db.QueryRowContext(ctx, "SELECT nextval('order_seq')").Scan(&seq); err != nil {
		return 0, fmt.Errorf("next order seq: %w", err)
	}
	return seq, nil
}

type scannerRow interface {
	Scan(dest ...interface{}) error
}

func scanOrder(row *sql.Row) (*entities.Order, error) {
	o, err := scanOrderRow(row)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNoRows
		}
		return nil, fmt.Errorf("get order: %w", err)
	}
	return o, nil
}

func scanOrderRow(row scannerRow) (*entities.Order, error) {
	o := &entities.Order{}
	var cancelledAt sql.NullTime
	err := row.Scan(
		&o.ID, &o.BuyerID, &o.SupplierID, &o.QuoteID, &o.RFQID,
		&o.Status, &o.PaymentStatus, &o.QCStatus,
		&o.TotalUSD,
		&o.ShippingAddress, &o.ShippingCity, &o.ShippingCountry,
		&o.WarrantyUntil,
		&cancelledAt, &o.CancellationReason,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if cancelledAt.Valid {
		t := cancelledAt.Time
		o.CancelledAt = &t
	}
	return o, nil
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullTimePtr(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}
