-- Write your migrate up statements here
create table if not exists tag
(
    id         uuid primary key,
    name       text        not null unique check (length(name) <= 50),
    updated_at timestamptz not null default current_timestamp,
    created_at timestamptz not null default current_timestamp
);

---- create above / drop below ----
drop table if exists tag;
