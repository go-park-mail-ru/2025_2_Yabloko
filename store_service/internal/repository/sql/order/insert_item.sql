INSERT INTO order_item (id, order_id, store_item_id, price, quantity)
SELECT gen_random_uuid(), $1, si.id, si.price, ci.quantity
FROM cart_item ci
JOIN cart c on c.id = ci.cart_id
JOIN store_item si on si.id = ci.store_item_id
WHERE c.user_id = $2;