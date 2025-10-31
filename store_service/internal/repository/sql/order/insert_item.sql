insert into order_item (id, order_id, store_item_id, price, quantity)
select gen_random_uuid(), $1, si.id, si.price, ci.quantity
from cart_item ci
         join cart c on c.id = ci.cart_id
         join store_item si on si.id = ci.store_item_id
where c.user_id = $2;