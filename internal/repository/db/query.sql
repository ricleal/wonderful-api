-- name: ListUsers :many
WITH cursor_data AS (
    -- Get cursor reference data in a single query
    SELECT 
        CASE WHEN $2 != '' AND $2 IS NOT NULL THEN
            (SELECT registration FROM users WHERE id = $2)
        END as starting_after_registration,
        CASE WHEN $3 != '' AND $3 IS NOT NULL THEN
            (SELECT registration FROM users WHERE id = $3)
        END as ending_before_registration
)
SELECT
    u.id,
    u.name,
    u.email,
    u.phone,
    u.cell,
    u.picture,
    u.registration
FROM
    users u
CROSS JOIN cursor_data c
WHERE
    -- email substring
    (u.email LIKE '%' || $1 || '%' OR $1 IS NULL)
    -- starting_after - using composite comparison for efficiency
    AND ($2 = '' OR $2 IS NULL OR ( 
        (u.registration < c.starting_after_registration) OR 
        (u.registration = c.starting_after_registration AND u.id < $2)
    ))
    -- ending_before - using composite comparison for efficiency  
    AND ($3 = '' OR $3 IS NULL OR ( 
        (u.registration > c.ending_before_registration) OR 
        (u.registration = c.ending_before_registration AND u.id > $3)
    ))
ORDER BY
    u.registration DESC, u.id DESC
LIMIT $4;


-- name: LoadBulkUsers :copyfrom
INSERT INTO users (
    id,
    name,
    email,
    phone,
    cell,
    picture,
    registration
) VALUES (  
    $1, $2, $3, $4, $5, $6, $7
);

