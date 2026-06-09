-- order_events replaces tracking_events jsonb on orders. Every state transition,
-- inbound webhook, manual admin action, or document attachment is one row.
-- The UNIQUE constraint enforces idempotency for at-least-once event sources.
CREATE TABLE order_events (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id     text NOT NULL REFERENCES orders(id),
    event_type   text NOT NULL,
    from_status  text,
    to_status    text,
    actor_id     text,                          -- soft ref to users(id) in user-service DB
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
    UNIQUE (order_id, event_type, occurred_at)   -- idempotency guard
);
