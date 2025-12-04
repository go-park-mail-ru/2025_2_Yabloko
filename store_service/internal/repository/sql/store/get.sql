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
        json_agg(DISTINCT st.tag_id) FILTER (
            WHERE
                st.tag_id IS NOT NULL
        ),
        '[]'::json
    ) AS tag_ids,
    COALESCE(
        json_agg(DISTINCT sc.category_id) FILTER (
            WHERE
                sc.category_id IS NOT NULL
        ),
        '[]'::json
    ) AS category_ids
FROM
    store s
    LEFT JOIN store_tag st ON st.store_id = s.id
    LEFT JOIN store_category sc ON sc.store_id = s.id
WHERE
    s.id = $1
GROUP BY
    s.id