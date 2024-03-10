-- name: ListUsers :many
SELECT
    id,
    name,
    email,
    phone,
    cell,
    picture,
    registration
FROM
    users
WHERE
	-- email substring
    (email LIKE '%' || $1 || '%' OR $1 IS NULL)
    -- starting_after
	AND ($2 = '' OR $2 IS NULL OR ( 
		(registration < (select registration from users where id = $2)) OR 
		(registration = (select registration from users where id = $2) AND id < $2)
	))
    -- ending_before
	AND ($3 = '' OR $3 IS NULL OR ( 
		(registration > (select registration from users where id = $3)) OR 
		(registration = (select registration from users where id = $3) AND id > $3)
	))
ORDER BY
    registration DESC, id DESC
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

