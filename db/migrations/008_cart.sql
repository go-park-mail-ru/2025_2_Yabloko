-- Write your migrate up statements here
create table if not exists cart
(
    id         uuid primary key,
    user_id    uuid        not null unique references account (id) on delete cascade,
    updated_at timestamptz not null default current_timestamp,
    created_at timestamptz not null default current_timestamp
);

create table if not exists cart_item
(
    id            uuid primary key,
    cart_id       uuid        not null references cart (id) on delete cascade,
    store_item_id uuid        not null references store_item (id) on delete cascade,
    quantity      int         not null check ( quantity >= 1 ),
    updated_at    timestamptz not null default current_timestamp,
    created_at    timestamptz not null default current_timestamp,
    unique (cart_id, store_item_id)
);
---- create above / drop below ----
drop table if exists cart;
drop table if exists cart_item;
