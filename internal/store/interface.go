package store

import (
	"context"

	"wonderful/internal/repository"
)

// Store is the interface that wraps the repositories.
type Store interface {
	Users() repository.UserRepository
	ExecTx(ctx context.Context, fn func(Store) error) error
}
