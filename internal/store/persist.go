package store

import (
	"context"

	"wonderful/internal/repository"
	"wonderful/internal/repository/db"
	"wonderful/internal/repository/db/sqlc"
)

// Store is a store for tweets and users.
type persistentStore struct {
	conn  sqlc.DBTX
	txMgr repository.TransactionManager
}

// NewPersistentStore creates a new store with the given database connection and transaction manager.
func NewPersistentStore(conn sqlc.DBTX, txMgr repository.TransactionManager) *persistentStore {
	return &persistentStore{
		conn:  conn,
		txMgr: txMgr,
	}
}

// Users returns a UserRepository for managing users.
func (s *persistentStore) Users() repository.UserRepository {
	return db.NewUserStorage(s.conn)
}

// Health returns a HealthRepository for database health checks.
func (s *persistentStore) Health() repository.HealthRepository {
	return db.NewHealthStorage(s.conn)
}

// ExecTx executes the given function within a database transaction.
func (s *persistentStore) ExecTx(ctx context.Context, fn func(Store) error) error {
	return s.txMgr.ExecTx(ctx, func(tx sqlc.DBTX) error {
		txStore := &persistentStore{
			conn:  tx,
			txMgr: s.txMgr, // Keep the same transaction manager
		}
		return fn(txStore)
	})
}
