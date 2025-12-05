SELECT
    si.id,
    i.name,
    si.price,
    i.description,
    i.card_img,
    COALESCE(
        json_agg(DISTINCT it.type_id) FILTER (WHERE it.type_id IS NOT NULL),
        '[]'::json
    ) AS type_ids
FROM
    store_item si
    INNER JOIN item i ON i.id = si.item_id
    LEFT JOIN item_type it ON it.item_id = i.id
WHERE
    si.store_id = $1
    AND (
        cardinality($2::uuid[]) = 0
        OR EXISTS (
            SELECT 1
            FROM item_type it2
            WHERE it2.item_id = i.id
              AND it2.type_id = ANY($2::uuid[])
        )
    )
GROUP BY
    si.id,
    i.name,
    si.price,
    i.description,
    i.card_img;
