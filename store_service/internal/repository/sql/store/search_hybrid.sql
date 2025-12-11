WITH bm25_search AS (
    SELECT
        s.id,
        ts_rank(s.search_vector, to_tsquery('russian', $1)) as bm25_score
    FROM
        store s
    WHERE
        s.search_vector @@ to_tsquery('russian', $1)
),
semantic_search AS (
    SELECT *
    FROM (
        SELECT
            s.id,
            1.0 - ((s.embedding <=> $2::vector) / 2.0) AS semantic_score
        FROM store s
        WHERE s.embedding IS NOT NULL
    ) t
    WHERE t.semantic_score > 0.7
    ORDER BY t.semantic_score DESC
    LIMIT 100
),
combined_results AS (
    SELECT
        COALESCE(b.id, s.id) as store_id,
        COALESCE(b.bm25_score, 0) * $3::float8 + COALESCE(s.semantic_score, 0) * $4::float8 as combined_score
    FROM
        bm25_search b 
    FULL OUTER JOIN semantic_search s ON b.id = s.id
)
SELECT
    s.id,
    s.name,
    s.description,
    s.city_id,
    s.address,
    s.card_img,
    s.rating,
    s.open_at,
    s.closed_at,
    COALESCE(
        json_agg(DISTINCT st.tag_id) FILTER (
            WHERE
                st.tag_id IS NOT NULL
        ),
        '[]'::json
    ) AS tag_ids,
    COALESCE(
        json_agg(DISTINCT sc.category_id) FILTER (
            WHERE
                sc.category_id IS NOT NULL
        ),
        '[]'::json
    ) AS category_ids,
    si.id AS item_id,
    i.name AS item_name,
    si.price,
    i.card_img AS item_card_img,
    i.embedding AS item_embedding,
    COALESCE(
        json_agg(DISTINCT it.type_id) FILTER (
            WHERE
                it.type_id IS NOT NULL
        ),
        '[]'::json
    ) AS item_types
FROM
    combined_results cr
    JOIN store s ON s.id = cr.store_id
    LEFT JOIN store_tag st ON s.id = st.store_id
    LEFT JOIN store_category sc ON s.id = sc.store_id
    LEFT JOIN store_item si ON s.id = si.store_id
    LEFT JOIN item i ON si.item_id = i.id
    LEFT JOIN item_type it ON i.id = it.item_id
GROUP BY s.id, s.name, s.description, s.city_id, s.address, s.card_img, s.rating, s.open_at, s.closed_at,
         si.id, i.name, si.price, i.card_img, i.embedding, cr.combined_score
ORDER BY cr.combined_score DESC
