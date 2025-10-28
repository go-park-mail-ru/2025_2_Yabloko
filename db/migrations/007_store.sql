-- Write your migrate up statements here
create table if not exists store
(
    id          uuid primary key,
    name        text        not null check (length(name) <= 50),
    description text        not null check (length(description) >= 30 and length(description) <= 2000),
    city_id     uuid        references city (id) on delete set null,
    address     text        not null check (length(address) <= 200),
    card_img    text check (card_img ~ '^(?:\.?/[\w.-]+)+\.(png|jpg|jpeg|svg)$'),
    rating      numeric(2, 1) check ( rating >= 0 and rating <= 5 ),
    open_at     timetz      not null,
    closed_at   timetz      not null,
    updated_at  timestamptz not null default current_timestamp,
    created_at  timestamptz not null default current_timestamp,
    unique (name, city_id, address)
);

CREATE TRIGGER trg_update_store_updated_at
    BEFORE UPDATE
    ON store
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

create table store_tag
(
    id         uuid primary key,
    store_id   uuid        not null references store (id) on delete cascade,
    tag_id     uuid        not null references tag (id) on delete cascade,
    updated_at timestamptz not null default current_timestamp,
    created_at timestamptz not null default current_timestamp,
    unique (store_id, tag_id)
);

CREATE TRIGGER trg_update_store_tag_updated_at
    BEFORE UPDATE
    ON store_tag
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

create table store_category
(
    id          uuid primary key,
    store_id    uuid        not null references store (id) on delete cascade,
    category_id uuid        not null references category (id) on delete cascade,
    updated_at  timestamptz not null default current_timestamp,
    created_at  timestamptz not null default current_timestamp,
    unique (store_id, category_id)
);

CREATE TRIGGER trg_update_store_category_updated_at
    BEFORE UPDATE
    ON store_category
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

create table if not exists store_item
(
    id         uuid primary key,
    store_id   uuid          not null references store (id) on delete cascade,
    item_id    uuid          not null references item (id) on delete cascade,
    price      numeric(8, 2) not null check ( price > 0 ),
    updated_at timestamptz   not null default current_timestamp,
    created_at timestamptz   not null default current_timestamp,
    unique (store_id, item_id)
);

CREATE TRIGGER trg_update_store_item_updated_at
    BEFORE UPDATE
    ON store_item
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

---- create above / drop below ----

drop table if exists store;
drop table if exists store_tag;
drop table if exists store_category;
drop table if exists store_item;
