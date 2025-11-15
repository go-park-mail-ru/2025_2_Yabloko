select
    si.id as id,
    it.name as name,
    it.card_img as card_img,
    si.price as price,
    ci.quantity as quantity
from
    cart c
    join cart_item ci on ci.cart_id = c.id
    join store_item si on si.id = ci.store_item_id
    join item it on it.id = si.item_id
where
    c.user_id = $1
order by
    ci.created_at;