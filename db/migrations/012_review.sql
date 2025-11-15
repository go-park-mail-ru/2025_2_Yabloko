-- Write your migrate up statements here
create table if not exists review
(
    id         uuid primary key,
    user_id    uuid          references account (id) on delete set null,
    store_id   uuid          not null references store (id) on delete cascade,
    rating     numeric(2, 1) not null check ( rating >= 0 and rating <= 5 ),
    comment    text check ( length(comment) <= 5000),
    updated_at timestamptz   not null default current_timestamp,
    created_at timestamptz   not null default current_timestamp
);


CREATE TRIGGER trg_update_review_updated_at
    BEFORE UPDATE
    ON review
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

---- create above / drop below ----
drop table if exists review;
