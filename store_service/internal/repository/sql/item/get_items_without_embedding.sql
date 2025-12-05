SELECT
    id,
    name,
    description
FROM
    item
WHERE
    embedding IS NULL
LIMIT
    1000