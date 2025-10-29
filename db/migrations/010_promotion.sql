-- Write your migrate up statements here
create table if not exists promotion
(
    id                uuid primary key,
    name              text        not null check (length(name) <= 50),
    relative_discount numeric(4, 2) check ( relative_discount >= 0 ) check ( relative_discount <= 99 ) default 0,
    absolute_discount numeric(8, 2) check ( absolute_discount >= 0 )                                   default 0,
    start_at          timestamptz not null,
    end_at            timestamptz not null check ( end_at > start_at ),
    updated_at        timestamptz not null                                                             default current_timestamp,
    created_at        timestamptz not null                                                             default current_timestamp,
    constraint only_one_discount check ((relative_discount > 0 and absolute_discount = 0) or
                                        (absolute_discount > 0 and relative_discount = 0))
);

CREATE TRIGGER trg_update_promotion_updated_at
    BEFORE UPDATE
    ON promotion
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

create table if not exists promotion_item
(
    id           uuid primary key,
    item_id      uuid        not null references item (id) on delete cascade,
    promotion_id uuid        not null references promotion (id) on delete cascade,
    updated_at   timestamptz not null default current_timestamp,
    created_at   timestamptz not null default current_timestamp,
    unique (item_id, promotion_id)
);

CREATE TRIGGER trg_update_promotion_item_updated_at
    BEFORE UPDATE
    ON promotion_item
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

---- create above / drop below ----
drop table if exists promotion;
drop table if exists promotion_item;
