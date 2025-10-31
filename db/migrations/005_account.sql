-- Write your migrate up statements here
create table if not exists account
(
    id         uuid primary key,
    email      text        not null unique check ( length(email) <= 100),
    hash       text        not null,
    name       text check (length(name) <= 60),
    phone      text,
    city_id uuid references city (id) on delete set null,
    address    text check (length(address) <= 200),
    updated_at timestamptz not null default current_timestamp,
    created_at timestamptz not null default current_timestamp,
    constraint email_format check (email ~ '^[a-zA-Z0-9._-]+@[a-zA-Z0-9._-]+\.[a-zA-Z0-9_-]+$'),
    constraint phone_length_min check (length(phone) >= 10),
    constraint phone_length_max check (length(phone) <= 16),
    constraint phone_format check (phone ~ '^[0-9]+$')
);

CREATE TRIGGER trg_update_account_updated_at
    BEFORE UPDATE
    ON account
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

create table if not exists friend
(
    id        uuid primary key,
    user_id_1 uuid not null references account (id) on delete cascade,
    user_id_2 uuid not null references account (id) on delete cascade,
    updated_at timestamptz not null default current_timestamp,
    created_at timestamptz not null default current_timestamp,
    constraint no_self_friend check (user_id_1 <> user_id_2),
    unique (user_id_1, user_id_2)
);

CREATE TRIGGER trg_update_friend_updated_at
    BEFORE UPDATE
    ON friend
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

---- create above / drop below ----
drop table if exists account;
drop table if exists friend;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.