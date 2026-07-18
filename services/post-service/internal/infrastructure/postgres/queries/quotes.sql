-- name: NextQuoteSeq :one
SELECT nextval('quote_id_seq')::bigint;

-- name: CreatePendingQuote :one
INSERT INTO quotes (id, rfq_id, supplier_id, product_id, match_score, status)
VALUES ($1, $2, $3, $4, $5, 'pending')
RETURNING *;

-- name: ListQuotesForRFQ :many
SELECT * FROM quotes
WHERE rfq_id = $1
ORDER BY created_at ASC;

-- name: GetQuoteForSupplier :one
SELECT * FROM quotes
WHERE rfq_id = $1 AND supplier_id = $2
ORDER BY created_at DESC
LIMIT 1;

-- name: SubmitQuote :one
-- The pending row was created with the RFQ; submission fills in commercial
-- terms. Guarding on status = 'pending' makes re-submission a no-rows error
-- instead of silently overwriting an accepted/rejected quote.
UPDATE quotes
SET price_usd      = $3,
    lead_time_days = $4,
    validity_date  = $5,
    supplier_notes = $6,
    status         = 'submitted',
    submitted_at   = now()
WHERE rfq_id = $1 AND supplier_id = $2 AND status = 'pending'
RETURNING *;

-- name: InsertManufacturerQuote :one
INSERT INTO quotes (id, rfq_id, manufacturer_id, product_id, price_usd, lead_time_days, validity_date, supplier_notes, status, submitted_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'submitted', now())
RETURNING *;

-- name: GetQuoteByID :one
SELECT * FROM quotes WHERE id = $1;

-- name: AcceptQuote :one
UPDATE quotes SET status = 'accepted'
WHERE id = $1 AND rfq_id = $2 AND status = 'submitted'
RETURNING *;

-- name: RejectOtherQuotes :exec
UPDATE quotes SET status = 'rejected'
WHERE rfq_id = $1 AND id <> $2 AND status IN ('pending', 'submitted');
