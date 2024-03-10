package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"wonderful/internal/repository"
	"wonderful/internal/repository/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/segmentio/ksuid"
)

// UserStorage is a postgres implementation of the repository.UserStorage interface.
type UserStorage struct {
	queries *sqlc.Queries
}

// NewUserStorage returns a new UserServer.
func NewUserStorage(dbConn sqlc.DBTX) *UserStorage {
	queries := sqlc.New(dbConn)
	return &UserStorage{
		queries: queries,
	}
}

func formatParameters(p repository.Params) sqlc.ListUsersParams {
	params := sqlc.ListUsersParams{}
	// This is for safety. The API by default returns a limit of 10.
	if p.Limit == 0 {
		p.Limit = 10
	}
	params.Limit = int32(p.Limit)
	// Email
	params.Column1 = pgtype.Text{}
	if p.Email != nil {
		params.Column1 = pgtype.Text{String: *p.Email, Valid: true}
	}
	// StartingAfter
	params.Column2 = pgtype.Text{}
	if p.StartingAfter != nil {
		params.Column2 = pgtype.Text{String: p.StartingAfter.String(), Valid: true}
	}
	// EndingBefore
	params.Column3 = pgtype.Text{}
	if p.EndingBefore != nil {
		params.Column3 = pgtype.Text{String: p.EndingBefore.String(), Valid: true}
	}

	return params
}

// ListUsers returns a list of users.
func (s *UserStorage) ListUsers(ctx context.Context, p repository.Params) ([]repository.User, error) {
	params := formatParameters(p)

	rows, err := s.queries.ListUsers(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	users := make([]repository.User, 0, len(rows))
	for idx := range rows {
		// to avoid creating a new variable for each iteration, use a pointer to the current row
		r := rows[idx]
		var picture map[string]string
		// json unmarsal picture
		if err := json.Unmarshal(r.Picture, &picture); err != nil {
			// if there is an error, log it and continue to the next row
			slog.Error("failed to unmarshal picture", "error", err)
			continue
		}
		cell := ""
		if r.Cell.Valid {
			cell = r.Cell.String
		}
		id, err := ksuid.Parse(r.ID)
		if err != nil {
			// if there is an error, log it and continue to the next row
			slog.Error("failed to parse id", "error", err)
			continue
		}
		users = append(users, repository.User{
			ID:           id,
			Name:         r.Name,
			Email:        r.Email,
			Phone:        r.Phone,
			Cell:         cell,
			Picture:      picture,
			Registration: r.Registration.Time,
		})
	}
	return users, nil
}

// Create creates multiple users.
func (s *UserStorage) Create(ctx context.Context, users []repository.User) error {
	params := make([]sqlc.LoadBulkUsersParams, 0, len(users))
	for _, u := range users {
		// json marsal picture to byte
		picture, err := json.Marshal(u.Picture)
		if err != nil {
			// if there is an error, log it and continue to the next user
			slog.Error("failed to marshal picture", "error", err)
			continue
		}
		var cell pgtype.Text
		if u.Cell != "" {
			cell = pgtype.Text{String: u.Cell, Valid: true}
		}
		params = append(params, sqlc.LoadBulkUsersParams{
			ID:           ksuid.New().String(),
			Name:         u.Name,
			Email:        u.Email,
			Phone:        u.Phone,
			Cell:         cell,
			Picture:      picture,
			Registration: pgtype.Timestamp{Time: u.Registration, Valid: true},
		})
	}

	_, err := s.queries.LoadBulkUsers(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create users: %w", err)
	}
	return nil
}
