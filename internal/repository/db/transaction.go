package db

import (
	"context"
	"fmt"

	"wonderful/internal/repository"
	"wonderful/internal/repository/db/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

// transactionManager implements the TransactionManager interface.
type transactionManager struct {
	pool *pgxpool.Pool
}

// NewTransactionManager creates a new transaction manager with the given connection pool.
func NewTransactionManager(pool *pgxpool.Pool) repository.TransactionManager {
	return &transactionManager{
		pool: pool,
	}
}

// ExecTx executes the given function within a database transaction.
func (tm *transactionManager) ExecTx(ctx context.Context, fn func(sqlc.DBTX) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("BeginTx: %w", err)
	}

	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("ExecTx: %w: Rollback: %w", err, rbErr)
		}
		return fmt.Errorf("ExecTx: %w", err)
	}

	return tx.Commit(ctx) //nolint:wrapcheck //no need to wrap here
}
