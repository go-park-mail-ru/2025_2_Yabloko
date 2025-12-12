create table if not exists promocode
(
    id                 uuid primary key,
    code               text        not null check (length(code) < 32),
    relative_discount  numeric(4, 2) check ( relative_discount >= 0 ) check ( relative_discount <= 99 ) default 0,
    absolute_discount  numeric(8, 2) check ( absolute_discount >= 0 )                                  default 0,
    start_at           timestamptz not null,
    end_at             timestamptz not null check ( end_at > start_at ),
    updated_at         timestamptz not null default current_timestamp,
    created_at         timestamptz not null default current_timestamp,
    check ( (relative_discount = 0 and absolute_discount > 0)
         or (relative_discount > 0 and absolute_discount = 0) )
);

CREATE TRIGGER trg_update_promocode_updated_at
    BEFORE UPDATE
    ON promocode
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

create table if not exists promocode_account
(
    id           uuid primary key,
    user_id      uuid        not null references account (id) on delete cascade,
    promocode_id uuid        not null references promocode (id) on delete cascade,
    updated_at   timestamptz not null default current_timestamp,
    created_at   timestamptz not null default current_timestamp,
    unique (user_id, promocode_id)
);

CREATE TRIGGER trg_update_promocode_account_updated_at
    BEFORE UPDATE
    ON promocode_account
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

---- create above / drop below ----
drop table if exists promocode_account;
drop table if exists promocode;
