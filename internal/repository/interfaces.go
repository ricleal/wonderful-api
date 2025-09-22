package repository

import (
	"context"

	"wonderful/internal/repository/db/sqlc"
)

// UserRepository represents a repository for users.
type UserRepository interface {
	ListUsers(ctx context.Context, p Params) ([]User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	Create(ctx context.Context, users []User) error
}

// HealthRepository represents a repository for database health checks.
type HealthRepository interface {
	Ping(ctx context.Context) error
	GetConnectionStats() map[string]interface{}
}

// TransactionManager handles database transactions.
type TransactionManager interface {
	ExecTx(ctx context.Context, fn func(sqlc.DBTX) error) error
}
