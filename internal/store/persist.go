package store

import (
	"context"
	"errors"
	"fmt"

	"wonderful/internal/repository"
	"wonderful/internal/repository/db"
	"wonderful/internal/repository/db/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store is a store for tweets and users.
type persistentStore struct {
	conn sqlc.DBTX
}

// NewPersistentStore creates a new store with the given database connection.
func NewPersistentStore(conn sqlc.DBTX) *persistentStore {
	return &persistentStore{
		conn: conn,
	}
}

// Users returns a UserRepository for managing users.
func (s *persistentStore) Users() repository.UserRepository {
	return db.NewUserStorage(s.conn)
}

// ExecTx executes the given function within a database transaction.
// See the test file for an example of how to use this function.
func (s *persistentStore) ExecTx(ctx context.Context, fn func(Store) error) error {
	conn, ok := s.conn.(*pgxpool.Pool)
	if !ok {
		return errors.New("ExecTx: db is not a *sql.DB")
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("BeginTx: %w", err)
	}
	newStore := NewPersistentStore(tx)
	err = fn(newStore)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("ExecTx: %w: Rollback: %w", err, rbErr)
		}
		return fmt.Errorf("ExecTx: %w", err)
	}
	return tx.Commit(ctx) //nolint:wrapcheck //no need to wrap here
}
