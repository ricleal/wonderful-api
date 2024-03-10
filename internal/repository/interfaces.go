package repository

import (
	"context"
)

// UserRepository represents a repository for users.
type UserRepository interface {
	ListUsers(ctx context.Context, p Params) ([]User, error)
	Create(ctx context.Context, users []User) error
}
