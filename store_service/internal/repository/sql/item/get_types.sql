SELECT
    DISTINCT t.id,
    t.name
FROM
    type t
    INNER JOIN item_type it ON it.type_id = t.id
    INNER JOIN store_item si ON si.item_id = it.item_id
WHERE
    si.store_id = $1
ORDER BY
    t.name ASC