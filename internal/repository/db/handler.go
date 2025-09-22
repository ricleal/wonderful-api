package db

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

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

	// Configure connection pool settings
	configureConnectionPool(config)

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to establish connection: %w", err)
	}

	// Test the connection with retry logic
	if err := testConnectionWithRetry(ctx, pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connection health check failed: %w", err)
	}

	logConnectionStatus(pool)

	return &Storage{
		pool: pool,
	}, nil
}

// configureConnectionPool sets up reasonable defaults for the connection pool
func configureConnectionPool(config *pgxpool.Config) {
	// Set maximum number of connections in the pool
	if maxConns := os.Getenv("DB_MAX_CONNECTIONS"); maxConns != "" {
		if val, err := strconv.Atoi(maxConns); err == nil && val > 0 {
			config.MaxConns = int32(val)
		}
	} else {
		config.MaxConns = 25 // Default reasonable maximum
	}

	// Set minimum number of connections in the pool
	if minConns := os.Getenv("DB_MIN_CONNECTIONS"); minConns != "" {
		if val, err := strconv.Atoi(minConns); err == nil && val >= 0 {
			config.MinConns = int32(val)
		}
	} else {
		config.MinConns = 5 // Default minimum to keep warm connections
	}

	// Set maximum connection lifetime
	if maxLifetime := os.Getenv("DB_MAX_CONN_LIFETIME"); maxLifetime != "" {
		if val, err := time.ParseDuration(maxLifetime); err == nil {
			config.MaxConnLifetime = val
		}
	} else {
		config.MaxConnLifetime = time.Hour // Default 1 hour
	}

	// Set maximum connection idle time
	if maxIdleTime := os.Getenv("DB_MAX_CONN_IDLE_TIME"); maxIdleTime != "" {
		if val, err := time.ParseDuration(maxIdleTime); err == nil {
			config.MaxConnIdleTime = val
		}
	} else {
		config.MaxConnIdleTime = 30 * time.Minute // Default 30 minutes
	}

	// Set health check period
	if healthCheck := os.Getenv("DB_HEALTH_CHECK_PERIOD"); healthCheck != "" {
		if val, err := time.ParseDuration(healthCheck); err == nil {
			config.HealthCheckPeriod = val
		}
	} else {
		config.HealthCheckPeriod = time.Minute // Default 1 minute
	}
}

// testConnectionWithRetry tests the database connection with retry logic
func testConnectionWithRetry(ctx context.Context, pool *pgxpool.Pool) error {
	maxRetries := 3
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		if err := pool.Ping(ctx); err != nil {
			if i == maxRetries-1 {
				return fmt.Errorf("failed to ping database after %d attempts: %w", maxRetries, err)
			}
			slog.Warn("database ping failed, retrying", "attempt", i+1, "error", err)

			// Create timeout context derived from parent to respect cancellation
			sleepCtx, cancel := context.WithTimeout(ctx, retryDelay)
			select {
			case <-sleepCtx.Done():
			case <-ctx.Done():
				cancel()
				return ctx.Err()
			}
			cancel()

			retryDelay *= 2 // Exponential backoff
			continue
		}
		slog.Info("database connection established successfully", "attempt", i+1)
		return nil
	}
	return fmt.Errorf("unexpected error in retry loop")
}

// logConnectionStatus logs the current connection pool status
func logConnectionStatus(pool *pgxpool.Pool) {
	stat := pool.Stat()
	slog.Info("database connection pool initialized",
		"max_connections", stat.MaxConns(),
		"total_connections", stat.TotalConns(),
		"idle_connections", stat.IdleConns(),
		"acquired_connections", stat.AcquiredConns(),
	)
}

// GetConnectionStats returns current connection pool statistics
func (s *Storage) GetConnectionStats() map[string]interface{} {
	stat := s.pool.Stat()
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
