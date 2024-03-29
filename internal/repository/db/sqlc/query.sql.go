// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: query.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const listUsers = `-- name: ListUsers :many
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
LIMIT $4
`

type ListUsersParams struct {
	Column1 pgtype.Text
	Column2 interface{}
	Column3 interface{}
	Limit   int32
}

type ListUsersRow struct {
	ID           string
	Name         string
	Email        string
	Phone        string
	Cell         pgtype.Text
	Picture      []byte
	Registration pgtype.Timestamp
}

func (q *Queries) ListUsers(ctx context.Context, arg ListUsersParams) ([]ListUsersRow, error) {
	rows, err := q.db.Query(ctx, listUsers,
		arg.Column1,
		arg.Column2,
		arg.Column3,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListUsersRow
	for rows.Next() {
		var i ListUsersRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Email,
			&i.Phone,
			&i.Cell,
			&i.Picture,
			&i.Registration,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type LoadBulkUsersParams struct {
	ID           string
	Name         string
	Email        string
	Phone        string
	Cell         pgtype.Text
	Picture      []byte
	Registration pgtype.Timestamp
}
