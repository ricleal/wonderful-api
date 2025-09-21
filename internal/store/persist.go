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

// Ping checks if the database connection is alive
func (s *persistentStore) Ping(ctx context.Context) error {
	conn, ok := s.conn.(*pgxpool.Pool)
	if !ok {
		return errors.New("Ping: db is not a *pgxpool.Pool")
	}
	return conn.Ping(ctx)
}

// GetConnectionStats returns database connection statistics
func (s *persistentStore) GetConnectionStats() map[string]interface{} {
	conn, ok := s.conn.(*pgxpool.Pool)
	if !ok {
		return map[string]interface{}{
			"error": "connection is not a *pgxpool.Pool",
		}
	}

	stat := conn.Stat()
	return map[string]interface{}{
		"max_connections":      stat.MaxConns(),
		"total_connections":    stat.TotalConns(),
		"idle_connections":     stat.IdleConns(),
		"acquired_connections": stat.AcquiredConns(),
		"successful_acquires":  stat.AcquireCount(),
		"canceled_acquires":    stat.CanceledAcquireCount(),
		"acquire_duration":     stat.AcquireDuration(),
	}
}
