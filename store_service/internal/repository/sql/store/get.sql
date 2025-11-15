SELECT
    s.id,
    s.name,
    s.description,
    s.city_id,
    s.address,
    s.card_img,
    s.rating,
    s.open_at,
    s.closed_at,
    COALESCE(
        array_agg(st.tag_id) FILTER (
            WHERE
                st.tag_id IS NOT NULL
        ),
        '{}'
    ) AS tag_ids
FROM
    store s
    LEFT JOIN store_tag st ON st.store_id = s.id
WHERE
    s.id = $1
GROUP BY
    s.id,
    s.name,
    s.description,
    s.city_id,
    s.address,
    s.card_img,
    s.rating,
    s.open_at,
    s.closed_at