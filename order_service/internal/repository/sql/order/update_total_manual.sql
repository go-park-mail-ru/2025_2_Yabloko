UPDATE orders
SET total_price = $2
WHERE id = $1::uuid