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

CREATE INDEX IF NOT EXISTS idx_store_embedding_cosine ON store USING ivfflat (embedding vector_cosine_ops) WITH (lists = 50);

CREATE INDEX IF NOT EXISTS idx_item_embedding_cosine ON item USING ivfflat (embedding vector_cosine_ops) WITH (lists = 50);

---- create above / drop below ----
DROP INDEX IF EXISTS idx_item_embedding_cosine;

DROP INDEX IF EXISTS idx_store_embedding_cosine;

ALTER TABLE
    item DROP COLUMN IF EXISTS embedding;

ALTER TABLE
    store DROP COLUMN IF EXISTS embedding;

DROP EXTENSION IF EXISTS vector;