-- Write your migrate up statements here
create table if not exists account
(
    id         uuid primary key,
    email      text        not null unique,
    hash       text        not null,
    name       text,
    phone      text,
    city_id    uuid references city (id),
    address    text,
    avatar_url text check (length(avatar_url) <= 500),
    updated_at timestamptz not null default current_timestamp,
    created_at timestamptz not null default current_timestamp,
    constraint email_length check (length(email) <= 100),
    constraint email_format check (email ~ '^[a-zA-Z0-9._-]+@[a-zA-Z0-9._-]+\.[a-zA-Z0-9_-]+$'),
    constraint name_length check (length(name) <= 60),
    constraint phone_length_min check (length(phone) >= 10),
    constraint phone_length_max check (length(phone) <= 16),
    constraint phone_format check (phone ~ '^[0-9]+$'),
    constraint address check (length(address) <= 200)
);


---- create above / drop below ----
drop table if exists account;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.