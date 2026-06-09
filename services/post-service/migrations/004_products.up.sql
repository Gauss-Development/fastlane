CREATE TABLE products (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id     uuid NOT NULL REFERENCES suppliers(id),
    sku             text NOT NULL,
    name            text NOT NULL,
    name_zh         text,
    category        text NOT NULL,            -- 'transceiver', 'fiber', 'laser', ...
    specs           jsonb NOT NULL,           -- structured spec data
    price_usd       numeric(12,2),
    moq             int,
    stock_qty       int,
    lead_time_days  int,
    datasheet_url   text,
    embedding       vector(1024),             -- voyage-3 dimension
    created_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (supplier_id, sku)
);
