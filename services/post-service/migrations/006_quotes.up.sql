CREATE TABLE quotes (
    id              text PRIMARY KEY,                -- QUOTE-YYYYMMDD-NNNN-SZX
    rfq_id          text NOT NULL REFERENCES rfqs(id),
    supplier_id     uuid NOT NULL REFERENCES suppliers(id),
    product_id      uuid REFERENCES products(id),
    price_usd       numeric(12,2),
    lead_time_days  int,
    validity_date   date,
    supplier_notes  text,
    match_score     int,                             -- 0..100, AI-generated
    status          text NOT NULL DEFAULT 'pending'
                      CHECK (status IN ('pending','submitted','accepted','rejected')),
    submitted_at    timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now()
);
