-- name: NextRFQSeq :one
SELECT nextval('rfq_id_seq')::bigint;

-- name: CreateRFQ :one
INSERT INTO rfqs (
    id, buyer_id, buyer_email, buyer_company,
    query_text, parsed_specs, matched_product_ids,
    status, qty, target_date, shipping_address, notes, project_id
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7,
    $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: GetRFQByID :one
SELECT * FROM rfqs WHERE id = $1;

-- name: ListRFQsByBuyer :many
SELECT * FROM rfqs
WHERE buyer_id = $1
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')::text)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountRFQsByBuyer :one
SELECT COUNT(*)::int FROM rfqs
WHERE buyer_id = $1
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')::text);

-- name: ListOpenRFQs :many
SELECT * FROM rfqs
WHERE status = 'open'
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountOpenRFQs :one
SELECT COUNT(*)::int FROM rfqs WHERE status = 'open';

-- name: UpdateRFQStatus :one
UPDATE rfqs SET status = $2 WHERE id = $1
RETURNING *;

-- name: ListProductsByIDs :many
SELECT p.id, p.supplier_id, p.sku, p.name, p.name_zh, p.category,
       p.specs, p.price_usd, p.moq, p.stock_qty, p.lead_time_days, p.datasheet_url
FROM products p
WHERE p.id = ANY($1::uuid[]);

-- name: ListSuppliersByIDs :many
SELECT id, name, name_zh, city, contact_email, verified_at
FROM suppliers
WHERE id = ANY($1::uuid[]);
