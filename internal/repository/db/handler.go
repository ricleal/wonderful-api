package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Storage is struct that holds the database connection.
type Storage struct {
	pool *pgxpool.Pool
}

// Close closes the database connection.
func (s *Storage) Close() {
	s.pool.Close()
}

// DB returns the database connection.
func (s *Storage) Pool() *pgxpool.Pool {
	return s.pool
}

// NewStorage returns a new Handler with a database connection.
func NewStorage(ctx context.Context) (*Storage, error) {
	config, err := pgxpool.ParseConfig(os.Getenv("DB_URL"))
	if err != nil {
		return nil, fmt.Errorf("unable to parse configuration: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to establish connection: %w", err)
	}

	return &Storage{
		pool: pool,
	}, nil
}
