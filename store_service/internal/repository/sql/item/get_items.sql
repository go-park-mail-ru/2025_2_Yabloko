SELECT
    si.id,
    i.name,
    si.price,
    i.description,
    i.card_img,
    COALESCE(
        array_agg(DISTINCT it.type_id) FILTER (
            WHERE
                it.type_id IS NOT NULL
        ),
        '{}'
    ) AS type_ids
FROM
    store_item si
    INNER JOIN item i ON i.id = si.item_id
    LEFT JOIN item_type it ON it.item_id = i.id
WHERE
    si.store_id = $1
GROUP BY
    si.id,
    i.name,
    si.price,
    i.description,
    i.card_img
ORDER BY
    i.name ASC