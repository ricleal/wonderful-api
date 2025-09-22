package db

import (
	"context"
	"errors"

	"wonderful/internal/repository"
	"wonderful/internal/repository/db/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

// healthStorage implements the HealthRepository interface.
type healthStorage struct {
	conn sqlc.DBTX
}

// NewHealthStorage creates a new health storage with the given database connection.
func NewHealthStorage(conn sqlc.DBTX) repository.HealthRepository {
	return &healthStorage{
		conn: conn,
	}
}

// Ping checks if the database connection is alive
func (h *healthStorage) Ping(ctx context.Context) error {
	conn, ok := h.conn.(*pgxpool.Pool)
	if !ok {
		return errors.New("Ping: db is not a *pgxpool.Pool")
	}
	return conn.Ping(ctx)
}

// GetConnectionStats returns database connection statistics
func (h *healthStorage) GetConnectionStats() map[string]interface{} {
	conn, ok := h.conn.(*pgxpool.Pool)
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
