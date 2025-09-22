package security

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemorySecretStore(t *testing.T) {
	ctx := context.Background()
	config := &SecretConfig{
		DefaultTTL:      100 * time.Millisecond,
		RefreshInterval: 50 * time.Millisecond,
		MinSecretLength: 8,
		ValidateOnLoad:  true,
	}
	
	store := NewInMemorySecretStore(config)
	defer store.Close()

	t.Run("basic operations", func(t *testing.T) {
		key := "test_secret"
		value := "secret_value"
		
		// Test Set and Get
		err := store.Set(ctx, key, value, 0)
		require.NoError(t, err)
		
		retrieved, err := store.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, retrieved)
		
		// Test Exists
		assert.True(t, store.Exists(ctx, key))
		
		// Test TTL
		ttl, err := store.TTL(ctx, key)
		require.NoError(t, err)
		assert.Greater(t, ttl, time.Duration(0))
		
		// Test Delete
		err = store.Delete(ctx, key)
		require.NoError(t, err)
		
		assert.False(t, store.Exists(ctx, key))
		
		_, err = store.Get(ctx, key)
		assert.Error(t, err)
	})

	t.Run("expiration", func(t *testing.T) {
		key := "expiring_secret"
		value := "will_expire"
		
		// Set with very short TTL
		err := store.Set(ctx, key, value, 50*time.Millisecond)
		require.NoError(t, err)
		
		// Should exist immediately
		assert.True(t, store.Exists(ctx, key))
		
		// Wait for expiration
		time.Sleep(60 * time.Millisecond)
		
		// Should be expired
		assert.False(t, store.Exists(ctx, key))
		
		_, err = store.Get(ctx, key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("custom TTL", func(t *testing.T) {
		key := "custom_ttl_secret"
		value := "custom_value"
		customTTL := 200 * time.Millisecond
		
		err := store.Set(ctx, key, value, customTTL)
		require.NoError(t, err)
		
		ttl, err := store.TTL(ctx, key)
		require.NoError(t, err)
		
		// TTL should be approximately the custom value (allowing for execution time)
		assert.Greater(t, ttl, 150*time.Millisecond)
		assert.Less(t, ttl, customTTL+10*time.Millisecond)
	})
}

func TestEnvSecretProvider(t *testing.T) {
	ctx := context.Background()
	
	t.Run("without prefix", func(t *testing.T) {
		provider := NewEnvSecretProvider("")
		
		// Set up test environment variable
		testKey := "TEST_SECRET_123"
		testValue := "test_secret_value"
		os.Setenv(testKey, testValue)
		defer os.Unsetenv(testKey)
		
		// Test loading
		value, err := provider.LoadSecret(ctx, testKey)
		require.NoError(t, err)
		assert.Equal(t, testValue, value)
		
		// Test validation
		err = provider.Validate(ctx, testKey)
		require.NoError(t, err)
		
		// Test missing secret
		_, err = provider.LoadSecret(ctx, "NONEXISTENT_SECRET")
		assert.Error(t, err)
		
		err = provider.Validate(ctx, "NONEXISTENT_SECRET")
		assert.Error(t, err)
	})
	
	t.Run("with prefix", func(t *testing.T) {
		prefix := "MYAPP"
		provider := NewEnvSecretProvider(prefix)
		
		// Set up test environment variable with prefix
		envKey := "MYAPP_JWT_SECRET"
		testValue := "prefixed_secret_value"
		os.Setenv(envKey, testValue)
		defer os.Unsetenv(envKey)
		
		// Test loading with original key (provider should add prefix)
		value, err := provider.LoadSecret(ctx, "JWT_SECRET")
		require.NoError(t, err)
		assert.Equal(t, testValue, value)
	})
}

func TestDefaultSecretManager(t *testing.T) {
	ctx := context.Background()
	
	// Set up test JWT secret (avoiding weak patterns)
	testJWTSecret := "production-grade-jwt-signing-key-with-high-entropy-value"
	os.Setenv("JWT_SECRET", testJWTSecret)
	defer os.Unsetenv("JWT_SECRET")
	
	config := &SecretConfig{
		DefaultTTL:      1 * time.Second,
		RefreshInterval: 500 * time.Millisecond,
		MinSecretLength: 32,
		ValidateOnLoad:  false, // Disable validation for simpler testing
	}
	
	store := NewInMemorySecretStore(config)
	manager := NewDefaultSecretManager(store, config)
	defer manager.Close()

	t.Run("JWT secret retrieval", func(t *testing.T) {
		// First call should load from environment
		secret, err := manager.GetJWTSecret(ctx)
		require.NoError(t, err)
		assert.Equal(t, testJWTSecret, secret)
		
		// Second call should get from cache
		secret2, err := manager.GetJWTSecret(ctx)
		require.NoError(t, err)
		assert.Equal(t, testJWTSecret, secret2)
	})

	t.Run("secret validation", func(t *testing.T) {
		// Load the secret first
		_, err := manager.GetJWTSecret(ctx)
		require.NoError(t, err)
		
		// Validate it
		err = manager.ValidateSecret(ctx, "JWT_SECRET")
		require.NoError(t, err)
	})

	t.Run("secret refresh", func(t *testing.T) {
		// Load the secret first
		_, err := manager.GetJWTSecret(ctx)
		require.NoError(t, err)
		
		// Change the environment variable
		newSecret := "updated-production-grade-jwt-signing-key-with-high-entropy"
		os.Setenv("JWT_SECRET", newSecret)
		defer os.Setenv("JWT_SECRET", testJWTSecret) // Restore
		
		// Refresh should reload from environment
		err = manager.RefreshSecret(ctx, "JWT_SECRET")
		require.NoError(t, err)
		
		// Next get should return new value
		secret, err := manager.GetJWTSecret(ctx)
		require.NoError(t, err)
		assert.Equal(t, newSecret, secret)
	})
}

func TestSecretValidation(t *testing.T) {
	ctx := context.Background()
	
	t.Run("valid JWT secret", func(t *testing.T) {
		config := &SecretConfig{
			DefaultTTL:      1 * time.Second,
			RefreshInterval: 500 * time.Millisecond,
			MinSecretLength: 32,
			ValidateOnLoad:  true,
		}
		
		store := NewInMemorySecretStore(config)
		manager := NewDefaultSecretManager(store, config)
		defer manager.Close()
		
		validSecret := "production-grade-jwt-signing-key-with-extremely-high-entropy"
		os.Setenv("JWT_SECRET", validSecret)
		defer os.Unsetenv("JWT_SECRET")
		
		secret, err := manager.GetJWTSecret(ctx)
		require.NoError(t, err)
		assert.Equal(t, validSecret, secret)
	})

	t.Run("short JWT secret", func(t *testing.T) {
		config := &SecretConfig{
			DefaultTTL:      1 * time.Second,
			RefreshInterval: 500 * time.Millisecond,
			MinSecretLength: 32,
			ValidateOnLoad:  true,
		}
		
		store := NewInMemorySecretStore(config)
		manager := NewDefaultSecretManager(store, config)
		defer manager.Close()
		
		shortSecret := "short"
		os.Setenv("JWT_SECRET", shortSecret)
		defer os.Unsetenv("JWT_SECRET")
		
		_, err := manager.GetJWTSecret(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be at least")
	})

	t.Run("weak JWT secret", func(t *testing.T) {
		config := &SecretConfig{
			DefaultTTL:      1 * time.Second,
			RefreshInterval: 500 * time.Millisecond,
			MinSecretLength: 32,
			ValidateOnLoad:  true,
		}
		
		store := NewInMemorySecretStore(config)
		manager := NewDefaultSecretManager(store, config)
		defer manager.Close()
		
		weakSecret := "this-contains-the-word-secret-which-is-weak-but-long-enough"
		os.Setenv("JWT_SECRET", weakSecret)
		defer os.Unsetenv("JWT_SECRET")
		
		_, err := manager.GetJWTSecret(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "weak pattern")
	})
}

func TestSecretConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		config := DefaultSecretConfig()
		
		assert.Equal(t, 15*time.Minute, config.DefaultTTL)
		assert.Equal(t, 5*time.Minute, config.RefreshInterval)
		assert.Equal(t, 32, config.MinSecretLength)
		assert.True(t, config.ValidateOnLoad)
		
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("config validation", func(t *testing.T) {
		// Invalid TTL
		config := &SecretConfig{
			DefaultTTL:      -1 * time.Minute,
			RefreshInterval: 5 * time.Minute,
			MinSecretLength: 32,
			ValidateOnLoad:  true,
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TTL must be positive")
		
		// Invalid refresh interval
		config.DefaultTTL = 15 * time.Minute
		config.RefreshInterval = -1 * time.Minute
		err = config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "refresh interval must be positive")
		
		// Refresh interval >= TTL
		config.RefreshInterval = 20 * time.Minute
		err = config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "should be less than")
	})

	t.Run("load from env", func(t *testing.T) {
		// Set environment variables
		os.Setenv("SECRET_DEFAULT_TTL", "30m")
		os.Setenv("SECRET_REFRESH_INTERVAL", "10m")
		os.Setenv("SECRET_MIN_LENGTH", "64")
		os.Setenv("SECRET_VALIDATE_ON_LOAD", "false")
		
		defer func() {
			os.Unsetenv("SECRET_DEFAULT_TTL")
			os.Unsetenv("SECRET_REFRESH_INTERVAL")
			os.Unsetenv("SECRET_MIN_LENGTH")
			os.Unsetenv("SECRET_VALIDATE_ON_LOAD")
		}()
		
		config := LoadConfigFromEnv()
		
		assert.Equal(t, 30*time.Minute, config.DefaultTTL)
		assert.Equal(t, 10*time.Minute, config.RefreshInterval)
		assert.Equal(t, 64, config.MinSecretLength)
		assert.False(t, config.ValidateOnLoad)
	})
}