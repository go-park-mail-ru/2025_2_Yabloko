-- Write your migrate up statements here
create table if not exists review
(
    id         uuid primary key,
    user_id    uuid          references account (id) on delete set null,
    store_id   uuid          not null references store (id) on delete cascade,
    rating     numeric(2, 1) not null check (rating >= 0 and rating <= 5),
    comment    text          check (length(comment) <= 5000),
    updated_at timestamptz   not null default current_timestamp,
    created_at timestamptz   not null default current_timestamp,
    unique (user_id, store_id)
);

CREATE TRIGGER trg_update_review_updated_at
    BEFORE UPDATE
    ON review
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

create or replace function recalc_store_rating(p_store_id uuid)
returns void as
$$
begin
    update store s
    set rating = sub.avg_rating
    from (
        select
            r.store_id,
            coalesce(avg(r.rating), 0)::numeric(2,1) as avg_rating
        from review r
        where r.store_id = p_store_id
        group by r.store_id
    ) as sub
    where s.id = sub.store_id;
end;
$$ language plpgsql;

-- триггерная функция для пересчёта рейтинга при изменении отзывов
create or replace function trg_review_recalc_store_rating()
returns trigger as
$$
begin
    if tg_op in ('INSERT', 'UPDATE') then
        perform recalc_store_rating(NEW.store_id);
    end if;

    if tg_op in ('DELETE', 'UPDATE') then
        if tg_op = 'DELETE' or (tg_op = 'UPDATE' and OLD.store_id <> NEW.store_id) then
            perform recalc_store_rating(OLD.store_id);
        end if;
    end if;

    return null;
end;
$$ language plpgsql;

create trigger trg_review_recalc_store_rating
    after insert or update or delete
    on review
    for each row
execute function trg_review_recalc_store_rating();

---- create above / drop below ----
drop trigger if exists trg_review_recalc_store_rating on review;
drop function if exists trg_review_recalc_store_rating();
drop function if exists recalc_store_rating(uuid);
drop table if exists review;
