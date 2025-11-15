-- Write your migrate up statements here
CREATE OR REPLACE FUNCTION update_updated_at()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

create table if not exists tag
(
    id         uuid primary key,
    name       text        not null unique check (length(name) <= 50),
    updated_at timestamptz not null default current_timestamp,
    created_at timestamptz not null default current_timestamp
);

CREATE TRIGGER trg_update_tag_updated_at
    BEFORE UPDATE
    ON tag
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();
---- create above / drop below ----
drop table if exists tag;
