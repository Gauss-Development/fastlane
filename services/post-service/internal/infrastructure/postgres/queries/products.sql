-- name: CreateProduct :one
INSERT INTO products (
    supplier_id, sku, name, name_zh, category, specs,
    price_usd, moq, stock_qty, lead_time_days, datasheet_url
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10, $11
)
RETURNING *;

-- name: UpsertProduct :one
-- Idempotent on (supplier_id, sku); seed re-runs UPDATE in place. The existing
-- embedding is preserved on conflict so re-seeding doesn't invalidate the
-- already-computed vector (which is expensive to regenerate).
INSERT INTO products (
    supplier_id, sku, name, name_zh, category, specs,
    price_usd, moq, stock_qty, lead_time_days, datasheet_url
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10, $11
)
ON CONFLICT (supplier_id, sku) DO UPDATE SET
    name           = EXCLUDED.name,
    name_zh        = EXCLUDED.name_zh,
    category       = EXCLUDED.category,
    specs          = EXCLUDED.specs,
    price_usd      = EXCLUDED.price_usd,
    moq            = EXCLUDED.moq,
    stock_qty      = EXCLUDED.stock_qty,
    lead_time_days = EXCLUDED.lead_time_days,
    datasheet_url  = EXCLUDED.datasheet_url
RETURNING *;

-- name: GetProductByID :one
SELECT * FROM products WHERE id = $1;

-- name: GetProductBySupplierSKU :one
SELECT * FROM products WHERE supplier_id = $1 AND sku = $2 LIMIT 1;

-- name: ListProductsBySupplier :many
SELECT * FROM products
WHERE supplier_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListProductsMissingEmbedding :many
-- Used by cmd/embed CLI to find products that still need an embedding.
SELECT * FROM products
WHERE embedding IS NULL
ORDER BY created_at ASC
LIMIT $1;

-- name: ListProductsMissingEmbeddingWithSupplier :many
-- Same as above, but joins the supplier so the embedding-text builder has
-- everything it needs (supplier name, city, cluster) without an N+1 lookup.
SELECT
    p.id, p.supplier_id, p.sku, p.name, p.name_zh, p.category, p.specs,
    p.price_usd, p.moq, p.stock_qty, p.lead_time_days, p.datasheet_url,
    p.embedding, p.created_at,
    s.name    AS supplier_name,
    s.name_zh AS supplier_name_zh,
    s.city    AS supplier_city,
    s.cluster AS supplier_cluster
FROM products p
JOIN suppliers s ON s.id = p.supplier_id
WHERE p.embedding IS NULL
ORDER BY p.created_at ASC
LIMIT $1;

-- name: UpdateProductEmbedding :exec
UPDATE products
SET embedding = $2
WHERE id = $1;
