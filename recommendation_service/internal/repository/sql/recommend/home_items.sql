WITH recent_orders AS (
    SELECT o.id
    FROM "orders" o
    WHERE o.user_id = $1
      AND o.status IN ('paid', 'delivered')
    ORDER BY o.created_at DESC
    LIMIT 20
),
recent_items AS (
    SELECT DISTINCT oi.store_item_id
    FROM order_item oi
    JOIN recent_orders ro ON ro.id = oi.order_id
),
recent_item_embeddings AS (
    SELECT i.embedding
    FROM recent_items ri
    JOIN store_item si ON si.id = ri.store_item_id
    JOIN item i ON i.id = si.item_id
    WHERE i.embedding IS NOT NULL
),
user_profile_embedding AS (
    SELECT
        CASE
            WHEN COUNT(*) = 0 THEN NULL
            ELSE avg(embedding)
        END AS embedding
    FROM recent_item_embeddings
),
candidate_items AS (
    SELECT
        *
    FROM (
        SELECT
            i.id::text          AS id,
            si.store_id::text   AS store_id,
            i.name              AS name,
            si.price::float8    AS price,
            i.card_img          AS card_img,
            (1.0 - (up.embedding <=> i.embedding)) AS score
        FROM item i
        JOIN store_item si ON si.item_id = i.id
        CROSS JOIN user_profile_embedding up
        WHERE up.embedding IS NOT NULL
          AND i.embedding IS NOT NULL
          AND NOT EXISTS (
              SELECT 1
              FROM recent_items ri2
              WHERE ri2.store_item_id = si.id
          )

        UNION ALL

        SELECT
            i2.id::text         AS id,
            si2.store_id::text  AS store_id,
            i2.name             AS name,
            si2.price::float8   AS price,
            i2.card_img         AS card_img,
            0.0                 AS score
        FROM item i2
        JOIN store_item si2 ON si2.item_id = i2.id
        CROSS JOIN user_profile_embedding up2
        WHERE up2.embedding IS NULL
          AND i2.embedding IS NOT NULL
          AND NOT EXISTS (
              SELECT 1
              FROM recent_items ri3
              WHERE ri3.store_item_id = si2.id
          )
    ) AS unioned
    ORDER BY score DESC, price ASC, random()
    LIMIT $2
),
scored AS (
    SELECT
        id,
        store_id,
        name,
        price,
        card_img,
        score
    FROM candidate_items
)
SELECT
    id,
    store_id,
    name,
    price,
    card_img,
    COALESCE(score, 0)::float8 AS score
FROM scored;
