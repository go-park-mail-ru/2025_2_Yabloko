update "order"
set total_price = (select sum(si.price * ci.quantity)
                   from cart_item ci
                            join cart c on c.id = ci.cart_id
                            join store_item si on si.id = ci.store_item_id
                   where c.user_id = $1)
where id = $2;