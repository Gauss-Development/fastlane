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
