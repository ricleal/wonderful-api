package security

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CacheEntry represents a cached secret with expiration
type CacheEntry struct {
	Value     string
	ExpiresAt time.Time
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// TTL returns the remaining time to live
func (e *CacheEntry) TTL() time.Duration {
	if e.IsExpired() {
		return 0
	}
	return time.Until(e.ExpiresAt)
}

// InMemorySecretStore provides an in-memory implementation of SecretStore
type InMemorySecretStore struct {
	cache  sync.Map
	config *SecretConfig
}

// NewInMemorySecretStore creates a new in-memory secret store
func NewInMemorySecretStore(config *SecretConfig) *InMemorySecretStore {
	if config == nil {
		config = DefaultSecretConfig()
	}
	
	store := &InMemorySecretStore{
		config: config,
	}
	
	// Start cleanup goroutine to remove expired entries
	go store.cleanupExpired()
	
	return store
}

// Get retrieves a secret by key
func (s *InMemorySecretStore) Get(ctx context.Context, key string) (string, error) {
	value, exists := s.cache.Load(key)
	if !exists {
		return "", fmt.Errorf("secret not found: %s", key)
	}
	
	entry, ok := value.(*CacheEntry)
	if !ok {
		return "", fmt.Errorf("invalid cache entry for key: %s", key)
	}
	
	if entry.IsExpired() {
		s.cache.Delete(key)
		return "", fmt.Errorf("secret expired: %s", key)
	}
	
	return entry.Value, nil
}

// Set stores a secret with optional TTL
func (s *InMemorySecretStore) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = s.config.DefaultTTL
	}
	
	entry := &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
	
	s.cache.Store(key, entry)
	return nil
}

// Delete removes a secret
func (s *InMemorySecretStore) Delete(ctx context.Context, key string) error {
	s.cache.Delete(key)
	return nil
}

// Exists checks if a secret exists and is not expired
func (s *InMemorySecretStore) Exists(ctx context.Context, key string) bool {
	value, exists := s.cache.Load(key)
	if !exists {
		return false
	}
	
	entry, ok := value.(*CacheEntry)
	if !ok {
		return false
	}
	
	if entry.IsExpired() {
		s.cache.Delete(key)
		return false
	}
	
	return true
}

// TTL returns the remaining time to live for a secret
func (s *InMemorySecretStore) TTL(ctx context.Context, key string) (time.Duration, error) {
	value, exists := s.cache.Load(key)
	if !exists {
		return 0, fmt.Errorf("secret not found: %s", key)
	}
	
	entry, ok := value.(*CacheEntry)
	if !ok {
		return 0, fmt.Errorf("invalid cache entry for key: %s", key)
	}
	
	if entry.IsExpired() {
		s.cache.Delete(key)
		return 0, fmt.Errorf("secret expired: %s", key)
	}
	
	return entry.TTL(), nil
}

// Close cleans up resources
func (s *InMemorySecretStore) Close() error {
	// Clear all cached secrets
	s.cache.Range(func(key, value interface{}) bool {
		s.cache.Delete(key)
		return true
	})
	return nil
}

// cleanupExpired removes expired entries from the cache
func (s *InMemorySecretStore) cleanupExpired() {
	ticker := time.NewTicker(time.Minute) // Cleanup every minute
	defer ticker.Stop()
	
	for range ticker.C {
		s.cache.Range(func(key, value interface{}) bool {
			entry, ok := value.(*CacheEntry)
			if ok && entry.IsExpired() {
				s.cache.Delete(key)
			}
			return true
		})
	}
}