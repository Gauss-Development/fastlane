-- Catalog service database bootstrap.
-- pgvector enabled for consistency with the other Fiberlane DBs (room for future capability embeddings).
CREATE EXTENSION IF NOT EXISTS vector;
