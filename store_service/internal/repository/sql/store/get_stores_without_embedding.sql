SELECT
    id,
    name,
    description,
    city_id,
    address,
    card_img,
    rating,
    open_at,
    closed_at
FROM
    store
WHERE
    embedding IS NULL
LIMIT
    1000