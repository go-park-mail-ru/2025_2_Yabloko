UPDATE orders
SET total_price = (
    SELECT SUM(si.price * ci.quantity)
    FROM cart_item ci
    JOIN cart c on c.id = ci.cart_id
    JOIN store_item si on si.id = ci.store_item_id
    WHERE c.user_id = $1
)
WHERE id = $2;