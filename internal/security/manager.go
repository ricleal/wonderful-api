package security

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

// DefaultSecretManager provides a concrete implementation of SecretManager
type DefaultSecretManager struct {
	store     SecretStore
	providers map[string]SecretProvider
	config    *SecretConfig
}

// NewDefaultSecretManager creates a new secret manager with the given store and providers
func NewDefaultSecretManager(store SecretStore, config *SecretConfig) *DefaultSecretManager {
	if config == nil {
		config = DefaultSecretConfig()
	}
	
	manager := &DefaultSecretManager{
		store:     store,
		providers: make(map[string]SecretProvider),
		config:    config,
	}
	
	// Add default environment provider
	manager.AddProvider("env", NewEnvSecretProvider(""))
	
	return manager
}

// AddProvider adds a secret provider for a specific type
func (m *DefaultSecretManager) AddProvider(providerType string, provider SecretProvider) {
	m.providers[providerType] = provider
}

// GetJWTSecret retrieves the JWT signing secret
func (m *DefaultSecretManager) GetJWTSecret(ctx context.Context) (string, error) {
	return m.getSecret(ctx, "JWT_SECRET", "env")
}

// getSecret retrieves a secret with automatic caching and validation
func (m *DefaultSecretManager) getSecret(ctx context.Context, key, providerType string) (string, error) {
	// First, try to get from cache
	if secret, err := m.store.Get(ctx, key); err == nil {
		return secret, nil
	}
	
	// If not in cache or expired, load from provider
	provider, exists := m.providers[providerType]
	if !exists {
		return "", fmt.Errorf("provider not found: %s", providerType)
	}
	
	secret, err := provider.LoadSecret(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to load secret %s: %w", key, err)
	}
	
	// Validate secret if configured
	if m.config.ValidateOnLoad {
		if err := m.validateSecret(ctx, key, secret); err != nil {
			return "", fmt.Errorf("secret validation failed for %s: %w", key, err)
		}
	}
	
	// Cache the secret
	if err := m.store.Set(ctx, key, secret, m.config.DefaultTTL); err != nil {
		slog.Warn("failed to cache secret", "key", key, "error", err)
		// Don't fail if caching fails, just return the secret
	}
	
	return secret, nil
}

// RefreshSecret forces a refresh of a secret from its source
func (m *DefaultSecretManager) RefreshSecret(ctx context.Context, key string) error {
	// Remove from cache to force reload
	if err := m.store.Delete(ctx, key); err != nil {
		slog.Warn("failed to delete secret from cache during refresh", "key", key, "error", err)
	}
	
	// For JWT_SECRET, we know it comes from env provider
	providerType := "env"
	if strings.Contains(key, "JWT") {
		providerType = "env"
	}
	
	// Reload the secret
	_, err := m.getSecret(ctx, key, providerType)
	return err
}

// ValidateSecret checks if a secret meets security requirements
func (m *DefaultSecretManager) ValidateSecret(ctx context.Context, key string) error {
	secret, err := m.store.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("secret not found in cache: %w", err)
	}
	
	return m.validateSecret(ctx, key, secret)
}

// validateSecret performs actual validation logic
func (m *DefaultSecretManager) validateSecret(ctx context.Context, key, secret string) error {
	if secret == "" {
		return fmt.Errorf("secret cannot be empty")
	}
	
	// JWT secret specific validation
	if strings.Contains(key, "JWT") {
		if len(secret) < m.config.MinSecretLength {
			return fmt.Errorf("JWT secret must be at least %d characters long, got %d", 
				m.config.MinSecretLength, len(secret))
		}
		
		// Check for common weak secrets
		weakSecrets := []string{
			"secret", "password", "123456", "test", "dev", "development",
			"your-secret-key", "jwt-secret", "change-me",
		}
		
		secretLower := strings.ToLower(secret)
		for _, weak := range weakSecrets {
			if strings.Contains(secretLower, weak) {
				return fmt.Errorf("JWT secret contains weak pattern: %s", weak)
			}
		}
	}
	
	return nil
}

// Close cleans up resources
func (m *DefaultSecretManager) Close() error {
	return m.store.Close()
}

// NewSecretManager creates a properly configured SecretManager for the application
func NewSecretManager(config *SecretConfig) SecretManager {
	if config == nil {
		config = DefaultSecretConfig()
	}
	
	// Create in-memory store for now
	store := NewInMemorySecretStore(config)
	
	// Create manager with default providers
	manager := NewDefaultSecretManager(store, config)
	
	return manager
}