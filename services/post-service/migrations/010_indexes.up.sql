-- products: structured spec lookups + ANN search on embeddings.
CREATE INDEX IF NOT EXISTS idx_products_specs_gin
    ON products USING gin (specs);
CREATE INDEX IF NOT EXISTS idx_products_embedding_ivfflat
    ON products USING ivfflat (embedding vector_cosine_ops) WITH (lists = 50);
CREATE INDEX IF NOT EXISTS idx_products_supplier_id
    ON products (supplier_id);
CREATE INDEX IF NOT EXISTS idx_products_category
    ON products (category);

-- rfqs / quotes hot paths.
CREATE INDEX IF NOT EXISTS idx_rfqs_buyer_id
    ON rfqs (buyer_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_rfqs_status
    ON rfqs (status);
CREATE INDEX IF NOT EXISTS idx_quotes_rfq_id
    ON quotes (rfq_id);
CREATE INDEX IF NOT EXISTS idx_quotes_supplier_id
    ON quotes (supplier_id);
CREATE INDEX IF NOT EXISTS idx_quotes_status
    ON quotes (status);

-- orders dashboards + filters.
CREATE INDEX IF NOT EXISTS idx_orders_buyer_created
    ON orders (buyer_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_supplier_created
    ON orders (supplier_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_status
    ON orders (status);

-- order_events timeline + filters.
CREATE INDEX IF NOT EXISTS idx_order_events_order_occurred
    ON order_events (order_id, occurred_at);
CREATE INDEX IF NOT EXISTS idx_order_events_event_type
    ON order_events (event_type);
