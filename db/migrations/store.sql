-- Write your migrate up statements here
create table if not exists store (
    id          uuid primary key,
    name        text        not null,
    description text        not null,
    city_id     uuid        not null, --references city(id)
    address     text        not null,
    card_img    text,
    rating      numeric(2, 1),
    open_at     timetz      not null,
    closed_at   timetz      not null,
    updated_at  timestamptz not null default current_timestamp,
    created_at  timestamptz not null default current_timestamp,
    constraint name_length check (length(name) <= 30),
    constraint description_length_min check (length(description) >= 30),
    constraint description_length_max check (length(description) <= 2000),
    constraint address_length check (length(address) <= 200),
    constraint card_img_format check (card_img ~ '^(?:\.?/[\w.-]+)+\.(png|jpg|jpeg|svg)$'),
    constraint rating_min check (rating >= 0),
    constraint rating_max check (rating <= 5),
    constraint unique_store unique (name, city_id, address)
);
---- create above / drop below ----

drop table if exists store;