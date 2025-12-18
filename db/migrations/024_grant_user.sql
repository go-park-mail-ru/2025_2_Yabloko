GRANT CONNECT ON DATABASE postgres TO apple_user;

GRANT USAGE ON SCHEMA public TO apple_user;

GRANT SELECT, INSERT, UPDATE, DELETE ON 
    cart,
    cart_item,
    orders,
    order_item,
    promocode_account,
    review,
    payment
TO apple_user;

GRANT SELECT ON 
    store,
    item,
    store_item,
    tag,
    type,
    city,
    category,
    promocode,
    store_tag,
    store_category
TO apple_user;


GRANT INSERT, SELECT ON account TO apple_user;
GRANT UPDATE (avatar_url, name, hash, address, city_id, updated_at) 
ON account TO apple_user;


GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO apple_user;

GRANT USAGE ON SCHEMA pg_catalog TO apple_user;


GRANT EXECUTE ON FUNCTION 
    recalc_store_rating
TO apple_user;