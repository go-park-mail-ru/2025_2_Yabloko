-- Write your migrate up statements here
create table if not exists item (
    id uuid primary key,
    name text not null check (length(name) <= 50),
    description text not null check (length(description) <= 200),
    card_img text check (card_img ~ '\.(png|jpg|jpeg|svg|webp|gif)$'),
    updated_at timestamptz not null default current_timestamp,
    created_at timestamptz not null default current_timestamp
);

CREATE TRIGGER trg_update_item_updated_at
BEFORE UPDATE ON item
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE INDEX IF NOT EXISTS idx_item_name ON item (name);

create table if not exists item_type (
    id uuid primary key,
    item_id uuid not null references item (id) on delete cascade,
    type_id uuid not null references type (id) on delete cascade,
    updated_at timestamptz not null default current_timestamp,
    created_at timestamptz not null default current_timestamp,
    unique (item_id, type_id)
);

CREATE TRIGGER trg_update_item_type_updated_at
BEFORE UPDATE ON item_type
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE INDEX IF NOT EXISTS idx_item_type_item_id ON item_type (item_id);
CREATE INDEX IF NOT EXISTS idx_item_type_type_id ON item_type (type_id);

---- create above / drop below ----
drop table if exists item;
drop table if exists item_type;
