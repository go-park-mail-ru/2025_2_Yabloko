-- Write your migrate up statements here
CREATE EXTENSION IF NOT EXISTS vector;

ALTER TABLE
    store
ADD
    COLUMN IF NOT EXISTS embedding VECTOR(384);

ALTER TABLE
    item
ADD
    COLUMN IF NOT EXISTS embedding VECTOR(384);

ALTER TABLE
    store
ADD
    COLUMN IF NOT EXISTS search_vector tsvector GENERATED ALWAYS AS (
        to_tsvector('russian', name || ' ' || description)
    ) STORED;

CREATE INDEX IF NOT EXISTS idx_store_search_vector ON store USING GIN (search_vector);

CREATE INDEX IF NOT EXISTS idx_store_embedding_hnsw ON store USING hnsw (embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64);

ALTER TABLE
    item
ADD
    COLUMN IF NOT EXISTS search_vector tsvector GENERATED ALWAYS AS (to_tsvector('russian', name)) STORED;

CREATE INDEX IF NOT EXISTS idx_item_search_vector ON item USING GIN (search_vector);

CREATE INDEX IF NOT EXISTS idx_item_embedding_hnsw ON item USING hnsw (embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64);

---- create above / drop below ----
DROP INDEX IF EXISTS idx_item_embedding_hnsw;

DROP INDEX IF EXISTS idx_item_search_vector;

DROP INDEX IF EXISTS idx_store_embedding_hnsw;

DROP INDEX IF EXISTS idx_store_search_vector;

ALTER TABLE
    item DROP COLUMN IF EXISTS search_vector;

ALTER TABLE
    store DROP COLUMN IF EXISTS search_vector;

ALTER TABLE
    item DROP COLUMN IF EXISTS embedding;

ALTER TABLE
    store DROP COLUMN IF EXISTS embedding;

DROP EXTENSION IF EXISTS vector;