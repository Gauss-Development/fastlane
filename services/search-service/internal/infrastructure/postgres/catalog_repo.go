package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pgvector/pgvector-go"

	"search-service/internal/domain"
)

type CatalogRepo struct {
	db *sql.DB
}

func NewCatalogRepo(db *sql.DB) *CatalogRepo {
	return &CatalogRepo{db: db}
}

// vectorSearchQuery retrieves the nearest products by cosine distance, joined
// with supplier display fields. numeric is cast to float8 so database/sql can
// scan it directly. Spec filtering/boosting happens in the application layer on
// this candidate set (vector-first retrieval keeps the demo robust when spec
// extraction misses a field).
const vectorSearchQuery = `
SELECT p.id, p.supplier_id, p.sku, p.name, COALESCE(p.name_zh, ''), p.category,
       p.specs,
       COALESCE(p.price_usd, 0)::float8,
       COALESCE(p.moq, 0), COALESCE(p.stock_qty, 0), COALESCE(p.lead_time_days, 0),
       COALESCE(p.datasheet_url, ''),
       s.name, s.city, (s.verified_at IS NOT NULL),
       (p.embedding <=> $1)::float8 AS distance
FROM products p
JOIN suppliers s ON s.id = p.supplier_id
WHERE p.embedding IS NOT NULL
ORDER BY p.embedding <=> $1
LIMIT $2`

// VectorSearch returns up to `limit` candidate hits ordered by ascending cosine
// distance to queryVec.
func (r *CatalogRepo) VectorSearch(ctx context.Context, queryVec []float32, limit int) ([]domain.CatalogHit, error) {
	rows, err := r.db.QueryContext(ctx, vectorSearchQuery, pgvector.NewVector(queryVec), limit)
	if err != nil {
		return nil, fmt.Errorf("catalog vector search: %w", err)
	}
	defer rows.Close()

	var hits []domain.CatalogHit
	for rows.Next() {
		var h domain.CatalogHit
		if err := rows.Scan(
			&h.ID, &h.SupplierID, &h.SKU, &h.Name, &h.NameZh, &h.Category,
			&h.SpecsJSON,
			&h.PriceUSD,
			&h.MOQ, &h.StockQty, &h.LeadTimeDays,
			&h.DatasheetURL,
			&h.SupplierName, &h.SupplierCity, &h.SupplierVerified,
			&h.VectorDistance,
		); err != nil {
			return nil, fmt.Errorf("scan catalog hit: %w", err)
		}
		hits = append(hits, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate catalog hits: %w", err)
	}
	return hits, nil
}
