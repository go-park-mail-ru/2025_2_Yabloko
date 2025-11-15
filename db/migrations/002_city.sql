-- Write your migrate up statements here
create table if not exists city
(
    id         uuid primary key,
    name       text        not null unique check (length(name) <= 50),
    updated_at timestamptz not null default current_timestamp,
    created_at timestamptz not null default current_timestamp
);

CREATE TRIGGER trg_update_city_updated_at
    BEFORE UPDATE
    ON city
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

---- create above / drop below ----
drop table if exists city;
