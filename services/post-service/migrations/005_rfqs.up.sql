-- buyer_id is text rather than uuid REFERENCES users(id): the users table lives
-- in the user-service DB (postgres_user). Cross-service FKs are not possible;
-- buyer_id is a soft reference that the gateway / user-service guarantees.
CREATE TABLE rfqs (
    id                   text PRIMARY KEY,             -- RFQ-YYYYMMDD-NNNN-SZX
    buyer_id             text NOT NULL,                -- soft ref to users(id) in user-service DB
    query_text           text NOT NULL,
    parsed_specs         jsonb NOT NULL,
    matched_product_ids  uuid[] NOT NULL DEFAULT '{}',
    status               text NOT NULL DEFAULT 'open',
    qty                  int,
    target_date          date,
    shipping_address     text,
    notes                text,
    created_at           timestamptz NOT NULL DEFAULT now()
);
