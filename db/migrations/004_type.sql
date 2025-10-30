-- Write your migrate up statements here
create table if not exists type
(
    id         uuid primary key,
    name       text        not null unique check (length(name) <= 50),
    updated_at timestamptz not null default current_timestamp,
    created_at timestamptz not null default current_timestamp
);

CREATE TRIGGER trg_update_type_updated_at
    BEFORE UPDATE
    ON type
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();
---- create above / drop below ----
drop table if exists type;
