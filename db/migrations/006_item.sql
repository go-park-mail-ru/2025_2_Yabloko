-- Write your migrate up statements here
create table if not exists item
(
    id          uuid primary key,
    name        text        not null check (length(name) <= 50),
    description text        not null check ( length(description) <= 200),
    card_img    text check (card_img ~ '\.(png|jpg|jpeg|svg|webp|gif)$'),
    updated_at  timestamptz not null default current_timestamp,
    created_at  timestamptz not null default current_timestamp
);

CREATE TRIGGER trg_update_item_updated_at
    BEFORE UPDATE
    ON item
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

create table if not exists item_type
(
    id         uuid primary key,
    item_id uuid not null references item (id) on delete cascade,
    type_id uuid not null references type (id) on delete cascade,
    updated_at timestamptz not null default current_timestamp,
    created_at timestamptz not null default current_timestamp
);

CREATE TRIGGER trg_update_item_type_updated_at
    BEFORE UPDATE
    ON item_type
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();
---- create above / drop below ----
drop table if exists item;
drop table if exists item_type;
