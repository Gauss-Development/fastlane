CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE SEQUENCE IF NOT EXISTS order_seq;

CREATE TABLE IF NOT EXISTS orders (
    id                  text PRIMARY KEY,           -- ORD-YYYYMMDD-NNNN-SFO
    buyer_id            text NOT NULL,
    supplier_id         text NOT NULL,
    quote_id            text NOT NULL,
    rfq_id              text NOT NULL,

    status              text NOT NULL DEFAULT 'pending_payment'
                          CHECK (status IN (
                            'draft','pending_payment','paid','in_production',
                            'ready_for_qc','qc_in_progress','qc_failed','qc_passed',
                            'shipped_from_cn','in_transit','out_for_delivery',
                            'delivered','completed','cancelled','refunded','disputed'
                          )),
    payment_status      text NOT NULL DEFAULT 'unpaid'
                          CHECK (payment_status IN ('unpaid','paid','refunded')),
    qc_status           text
                          CHECK (qc_status IS NULL OR qc_status IN
                            ('pending','in_progress','passed','failed')),

    total_usd           numeric(12,2) NOT NULL,
    shipping_address    text,
    shipping_city       text,
    shipping_country    text,

    warranty_until      date,
    cancelled_at        timestamptz,
    cancellation_reason text,

    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),

    UNIQUE (quote_id)
);

CREATE INDEX IF NOT EXISTS orders_buyer_idx  ON orders (buyer_id);
CREATE INDEX IF NOT EXISTS orders_status_idx ON orders (status);

CREATE TABLE IF NOT EXISTS order_events (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id     text NOT NULL REFERENCES orders(id),
    event_type   text NOT NULL,
    from_status  text,
    to_status    text,
    actor_id     text,
    actor_type   text NOT NULL
                   CHECK (actor_type IN
                     ('buyer','supplier','admin','system','inspector','carrier')),
    occurred_at  timestamptz NOT NULL,
    occurred_tz  text NOT NULL DEFAULT 'UTC',
    location     text,
    payload      jsonb NOT NULL DEFAULT '{}',
    documents    jsonb NOT NULL DEFAULT '[]',
    notes        text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    UNIQUE (order_id, event_type, occurred_at)
);

CREATE INDEX IF NOT EXISTS order_events_order_idx ON order_events (order_id, occurred_at);
