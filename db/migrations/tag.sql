-- Write your migrate up statements here
create table tag (
    id uuid primary key,
    name text not null unique,
    updated_at timestamptz not null default current_timestamp,
    created_at timestamptz not null default current_timestamp,
    constraint name_length check (length(name) <= 50)
);

create table store_tag (
    store_id uuid not null references account(id),
    tag_id uuid not null references tag(id),
    updated_at timestamptz not null default current_timestamp,
    created_at timestamptz not null default current_timestamp,
    constraint unique_pair unique (store_id, tag_id)
);

---- create above / drop below ----
drop table store_tag;

drop table tag;
