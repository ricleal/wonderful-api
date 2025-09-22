package service

import (
	"context"

	"wonderful/internal/store"
)

// HealthService provides health check functionality
type HealthService interface {
	CheckDatabase(ctx context.Context) error
	GetDatabaseStats() map[string]interface{}
}

type healthService struct {
	store store.Store
}

// NewHealthService creates a new health service
func NewHealthService(store store.Store) HealthService {
	return &healthService{
		store: store,
	}
}

// CheckDatabase checks if the database is healthy
func (h *healthService) CheckDatabase(ctx context.Context) error {
	// Use the health repository to ping the database
	return h.store.Health().Ping(ctx)
}

// GetDatabaseStats returns database connection statistics
func (h *healthService) GetDatabaseStats() map[string]interface{} {
	return h.store.Health().GetConnectionStats()
}
