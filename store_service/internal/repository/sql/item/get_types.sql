SELECT DISTINCT type.id, type.name
FROM store_item
JOIN item_type ON store_item.item_id = item_type.item_id
JOIN type ON item_type.type_id = type.id
WHERE store_item.store_id = $1
ORDER BY type.name