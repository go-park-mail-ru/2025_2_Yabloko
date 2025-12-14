insert into review (id, user_id, store_id, rating, comment)
values (
    gen_random_uuid(),
    $1,
    $2,
    $3,
    $4
)
on conflict (user_id, store_id) do update
set rating     = excluded.rating,
    comment    = excluded.comment,
    updated_at = now();
