delete
from cart_item
where cart_id = (select id from cart where user_id = $1)
returning cart_id