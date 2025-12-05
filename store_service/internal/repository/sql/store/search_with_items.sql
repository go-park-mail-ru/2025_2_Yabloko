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
    ) AS category_ids,
    si.id AS item_id,
    i.name AS item_name,
    si.price,
    i.embedding AS item_embedding,
    COALESCE(
        json_agg(DISTINCT it.type_id) FILTER (
            WHERE
                it.type_id IS NOT NULL
        ),
        '[]'::json
    ) AS item_types
FROM
    store s
    LEFT JOIN store_tag st ON s.id = st.store_id
    LEFT JOIN store_category sc ON s.id = sc.store_id
    LEFT JOIN store_item si ON s.id = si.store_id
    LEFT JOIN item i ON si.item_id = i.id
    LEFT JOIN item_type it ON i.id = it.item_id