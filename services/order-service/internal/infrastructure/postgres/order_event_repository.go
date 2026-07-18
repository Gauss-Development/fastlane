package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"order-service/internal/domain/entities"
)

type OrderEventRepository struct {
	db *sql.DB
}

func NewOrderEventRepository(db *sql.DB) *OrderEventRepository {
	return &OrderEventRepository{db: db}
}

const eventCols = `
    id::text, order_id, event_type,
    COALESCE(from_status,''), COALESCE(to_status,''),
    COALESCE(actor_id,''), actor_type,
    occurred_at, COALESCE(occurred_tz,'UTC'),
    COALESCE(location,''), payload, documents,
    COALESCE(notes,''), created_at`

func (r *OrderEventRepository) Insert(ctx context.Context, e *entities.OrderEvent) (*entities.OrderEvent, error) {
	const q = `
INSERT INTO order_events (
    order_id, event_type, from_status, to_status,
    actor_id, actor_type,
    occurred_at, occurred_tz, location,
    payload, documents, notes
) VALUES (
    $1,$2,$3,$4,
    $5,$6,
    $7,$8,$9,
    $10,$11,$12
) ON CONFLICT (order_id, event_type, occurred_at) DO NOTHING
RETURNING ` + eventCols

	payload := e.Payload
	if len(payload) == 0 {
		payload = []byte("{}")
	}
	documents := e.Documents
	if len(documents) == 0 {
		documents = []byte("[]")
	}

	row := r.db.QueryRowContext(ctx, q,
		e.OrderID, e.EventType, nullStr(e.FromStatus), nullStr(e.ToStatus),
		nullStr(e.ActorID), e.ActorType,
		e.OccurredAt, e.OccurredTZ, nullStr(e.Location),
		payload, documents, nullStr(e.Notes),
	)
	out, err := scanEvent(row)
	if err != nil {
		return nil, fmt.Errorf("insert order event: %w", err)
	}
	return out, nil
}

func (r *OrderEventRepository) ListByOrder(ctx context.Context, orderID string) ([]*entities.OrderEvent, error) {
	const q = `SELECT ` + eventCols + ` FROM order_events WHERE order_id = $1 ORDER BY occurred_at ASC`
	rows, err := r.db.QueryContext(ctx, q, orderID)
	if err != nil {
		return nil, fmt.Errorf("list order events: %w", err)
	}
	defer rows.Close()

	var out []*entities.OrderEvent
	for rows.Next() {
		e, err := scanEventRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan order event: %w", err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order events: %w", err)
	}
	return out, nil
}

func scanEvent(row *sql.Row) (*entities.OrderEvent, error) {
	e, err := scanEventRow(row)
	if err != nil {
		if isNotFound(err) {
			// ON CONFLICT DO NOTHING returns no rows; treat as success with zero event
			return &entities.OrderEvent{}, nil
		}
		return nil, fmt.Errorf("get event: %w", err)
	}
	return e, nil
}

func scanEventRow(row scannerRow) (*entities.OrderEvent, error) {
	e := &entities.OrderEvent{}
	err := row.Scan(
		&e.ID, &e.OrderID, &e.EventType,
		&e.FromStatus, &e.ToStatus,
		&e.ActorID, &e.ActorType,
		&e.OccurredAt, &e.OccurredTZ,
		&e.Location, &e.Payload, &e.Documents,
		&e.Notes, &e.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return e, nil
}
