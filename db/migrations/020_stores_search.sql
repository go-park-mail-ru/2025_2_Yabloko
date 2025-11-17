ALTER TABLE store 
ADD COLUMN IF NOT EXISTS ts_name tsvector 
GENERATED ALWAYS AS (to_tsvector('russian', name)) STORED;

ALTER TABLE store 
ADD COLUMN IF NOT EXISTS ts_description tsvector 
GENERATED ALWAYS AS (to_tsvector('russian', description)) STORED;

ALTER TABLE store 
ADD COLUMN IF NOT EXISTS ts_search tsvector 
GENERATED ALWAYS AS (to_tsvector('russian', name || ' ' || description)) STORED;

CREATE INDEX IF NOT EXISTS idx_store_name_gin ON store USING GIN (ts_name);
CREATE INDEX IF NOT EXISTS idx_store_description_gin ON store USING GIN (ts_description);
CREATE INDEX IF NOT EXISTS idx_store_search_gin ON store USING GIN (ts_search);
