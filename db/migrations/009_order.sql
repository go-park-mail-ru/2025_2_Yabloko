-- Write your migrate up statements here
create type order_status as enum ('pending', 'paid', 'delivered', 'cancelled', 'on the way');

create table if not exists "order"
(
    id          uuid primary key,
    user_id     uuid         not null references account (id) on delete cascade,
    store_id    uuid         references store (id) on delete set null,
    total_price numeric(8, 2) check ( total_price > 0 ),
    status      order_status not null default 'pending',
    updated_at  timestamptz  not null default current_timestamp,
    created_at  timestamptz  not null default current_timestamp
);

create table if not exists order_item
(
    id            uuid primary key,
    order_id      uuid          not null references "order" (id) on delete cascade,
    store_item_id uuid          not null references store_item (id) on delete cascade,
    price         numeric(8, 2) not null check ( price > 0 ),
    quantity      int           not null check ( quantity >= 1 ),
    updated_at    timestamptz   not null default current_timestamp,
    created_at    timestamptz   not null default current_timestamp,
    unique (order_id, store_item_id)
);
---- create above / drop below ----
drop table if exists "order";
drop table if exists order_item;
drop type if exists order_status;
