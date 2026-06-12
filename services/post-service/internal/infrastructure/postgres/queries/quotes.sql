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
