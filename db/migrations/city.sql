-- Write your migrate up statements here
create table if not exists city (
    id         uuid primary key,
    name       text        not null unique,
    updated_at timestamptz not null default current_timestamp,
    created_at timestamptz not null default current_timestamp,
    constraint name_length check (length(name) <= 50)
);


---- create above / drop below ----
drop table if exists city;
