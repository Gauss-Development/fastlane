-- See Order Service Architecture doc for the full state machine.
-- buyer_id is text (soft ref to users in user-service DB).
CREATE TABLE orders (
    id                  text PRIMARY KEY,           -- ORD-YYYYMMDD-NNNN-SFO
    buyer_id            text NOT NULL,              -- soft ref to users(id) in user-service DB
    supplier_id         uuid NOT NULL REFERENCES suppliers(id),
    quote_id            text NOT NULL REFERENCES quotes(id),
    rfq_id              text NOT NULL REFERENCES rfqs(id),

    status              text NOT NULL DEFAULT 'draft'
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
    updated_at          timestamptz NOT NULL DEFAULT now()
);
