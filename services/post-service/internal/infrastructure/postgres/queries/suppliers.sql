-- name: CreateSupplier :one
INSERT INTO suppliers (
    name, name_zh, city, country, cluster,
    capabilities, certifications,
    founded_year, employees, facility_size_m2, annual_output,
    on_time_rate, rating, order_count,
    verified_at, audit_report_url, photo_url,
    contact_email, contact_wechat
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7,
    $8, $9, $10, $11,
    $12, $13, $14,
    $15, $16, $17,
    $18, $19
)
RETURNING *;

-- name: UpsertSupplier :one
-- Idempotent insert used by the seed CLI. Caller supplies a deterministic id
-- (uuid5 of supplier slug); re-runs UPDATE the row in place rather than
-- duplicating rows. order_count is preserved on UPDATE so seeding doesn't
-- clobber live counters.
INSERT INTO suppliers (
    id,
    name, name_zh, city, country, cluster,
    capabilities, certifications,
    founded_year, employees, facility_size_m2, annual_output,
    on_time_rate, rating,
    verified_at, audit_report_url, photo_url,
    contact_email, contact_wechat
) VALUES (
    $1,
    $2, $3, $4, $5, $6,
    $7, $8,
    $9, $10, $11, $12,
    $13, $14,
    $15, $16, $17,
    $18, $19
)
ON CONFLICT (id) DO UPDATE SET
    name              = EXCLUDED.name,
    name_zh           = EXCLUDED.name_zh,
    city              = EXCLUDED.city,
    country           = EXCLUDED.country,
    cluster           = EXCLUDED.cluster,
    capabilities      = EXCLUDED.capabilities,
    certifications    = EXCLUDED.certifications,
    founded_year      = EXCLUDED.founded_year,
    employees         = EXCLUDED.employees,
    facility_size_m2  = EXCLUDED.facility_size_m2,
    annual_output     = EXCLUDED.annual_output,
    on_time_rate      = EXCLUDED.on_time_rate,
    rating            = EXCLUDED.rating,
    verified_at       = EXCLUDED.verified_at,
    audit_report_url  = EXCLUDED.audit_report_url,
    photo_url         = EXCLUDED.photo_url,
    contact_email     = EXCLUDED.contact_email,
    contact_wechat    = EXCLUDED.contact_wechat
RETURNING *;

-- name: GetSupplierByID :one
SELECT * FROM suppliers WHERE id = $1;

-- name: GetSupplierByName :one
SELECT * FROM suppliers WHERE name = $1 LIMIT 1;

-- name: ListSuppliers :many
SELECT * FROM suppliers
WHERE
    (sqlc.narg('cluster')::text IS NULL OR cluster = sqlc.narg('cluster')::text)
    AND (sqlc.narg('verified_only')::boolean IS NOT TRUE OR verified_at IS NOT NULL)
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountSuppliers :one
SELECT COUNT(*)::int FROM suppliers
WHERE
    (sqlc.narg('cluster')::text IS NULL OR cluster = sqlc.narg('cluster')::text)
    AND (sqlc.narg('verified_only')::boolean IS NOT TRUE OR verified_at IS NOT NULL);
