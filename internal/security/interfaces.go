package security

import (
	"context"
	"time"
)

// SecretManager provides high-level secret management operations
type SecretManager interface {
	// GetJWTSecret retrieves the JWT signing secret
	GetJWTSecret(ctx context.Context) (string, error)
	
	// RefreshSecret forces a refresh of a secret from its source
	RefreshSecret(ctx context.Context, key string) error
	
	// ValidateSecret checks if a secret meets security requirements
	ValidateSecret(ctx context.Context, key string) error
	
	// Close cleans up resources
	Close() error
}

// SecretStore provides low-level secret storage and retrieval
// This interface can be implemented for different backends:
// - In-memory (current)
// - Redis (distributed)
// - HashiCorp Vault
// - AWS Secrets Manager
// - Kubernetes Secrets
type SecretStore interface {
	// Get retrieves a secret by key
	Get(ctx context.Context, key string) (string, error)
	
	// Set stores a secret with optional TTL
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	
	// Delete removes a secret
	Delete(ctx context.Context, key string) error
	
	// Exists checks if a secret exists
	Exists(ctx context.Context, key string) bool
	
	// TTL returns the remaining time to live for a secret
	TTL(ctx context.Context, key string) (time.Duration, error)
	
	// Close cleans up resources
	Close() error
}

// SecretProvider loads secrets from external sources
type SecretProvider interface {
	// LoadSecret loads a secret from its source (env, file, vault, etc.)
	LoadSecret(ctx context.Context, key string) (string, error)
	
	// Validate checks if the provider can handle the given key
	Validate(ctx context.Context, key string) error
}

// SecretConfig holds configuration for secret management
type SecretConfig struct {
	// DefaultTTL is the default time-to-live for cached secrets
	DefaultTTL time.Duration
	
	// RefreshInterval is how often to refresh secrets proactively
	RefreshInterval time.Duration
	
	// MinSecretLength is the minimum length for secrets
	MinSecretLength int
	
	// ValidateOnLoad determines if secrets should be validated when loaded
	ValidateOnLoad bool
}

// DefaultSecretConfig returns sensible defaults
func DefaultSecretConfig() *SecretConfig {
	return &SecretConfig{
		DefaultTTL:      15 * time.Minute, // Cache secrets for 15 minutes
		RefreshInterval: 5 * time.Minute,  // Refresh every 5 minutes
		MinSecretLength: 32,               // JWT secrets should be at least 32 chars
		ValidateOnLoad:  true,
	}
}