package security

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// EnvSecretProvider loads secrets from environment variables
type EnvSecretProvider struct {
	prefix string // Optional prefix for environment variables
}

// NewEnvSecretProvider creates a new environment variable secret provider
func NewEnvSecretProvider(prefix string) *EnvSecretProvider {
	return &EnvSecretProvider{
		prefix: prefix,
	}
}

// LoadSecret loads a secret from environment variables
func (p *EnvSecretProvider) LoadSecret(ctx context.Context, key string) (string, error) {
	envKey := key
	if p.prefix != "" {
		envKey = p.prefix + "_" + key
	}
	
	// Convert to uppercase for environment variable convention
	envKey = strings.ToUpper(envKey)
	
	value := os.Getenv(envKey)
	if value == "" {
		return "", fmt.Errorf("environment variable not found: %s", envKey)
	}
	
	return value, nil
}

// Validate checks if the provider can handle the given key
func (p *EnvSecretProvider) Validate(ctx context.Context, key string) error {
	envKey := key
	if p.prefix != "" {
		envKey = p.prefix + "_" + key
	}
	
	envKey = strings.ToUpper(envKey)
	
	if os.Getenv(envKey) == "" {
		return fmt.Errorf("environment variable not found: %s", envKey)
	}
	
	return nil
}