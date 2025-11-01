select type.id, type.name
from store_item
         join item_type on store_item.item_id = item_type.item_id
         join type on store_item.type_id = type.id
where store_item.store_id = $1