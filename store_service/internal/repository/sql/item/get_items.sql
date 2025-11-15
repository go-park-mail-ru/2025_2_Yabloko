select store_item.id, item.name, store_item.price, item.description, item.card_img, item_type.type_id
from store_item
         join item on store_item.item_id = item.id
         join item_type on item.id = item_type.item_id
where store_item.store_id = $1