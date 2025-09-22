package security

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// LoadConfigFromEnv loads SecretConfig from environment variables
func LoadConfigFromEnv() *SecretConfig {
	config := DefaultSecretConfig()
	
	// Load TTL configuration
	if ttlStr := os.Getenv("SECRET_DEFAULT_TTL"); ttlStr != "" {
		if ttl, err := time.ParseDuration(ttlStr); err == nil {
			config.DefaultTTL = ttl
		}
	}
	
	if refreshStr := os.Getenv("SECRET_REFRESH_INTERVAL"); refreshStr != "" {
		if refresh, err := time.ParseDuration(refreshStr); err == nil {
			config.RefreshInterval = refresh
		}
	}
	
	// Load validation configuration
	if minLenStr := os.Getenv("SECRET_MIN_LENGTH"); minLenStr != "" {
		if minLen, err := strconv.Atoi(minLenStr); err == nil && minLen > 0 {
			config.MinSecretLength = minLen
		}
	}
	
	if validateStr := os.Getenv("SECRET_VALIDATE_ON_LOAD"); validateStr != "" {
		if validate, err := strconv.ParseBool(validateStr); err == nil {
			config.ValidateOnLoad = validate
		}
	}
	
	return config
}

// Validate checks if the configuration is valid
func (c *SecretConfig) Validate() error {
	if c.DefaultTTL <= 0 {
		return fmt.Errorf("default TTL must be positive, got %v", c.DefaultTTL)
	}
	
	if c.RefreshInterval <= 0 {
		return fmt.Errorf("refresh interval must be positive, got %v", c.RefreshInterval)
	}
	
	if c.MinSecretLength <= 0 {
		return fmt.Errorf("minimum secret length must be positive, got %d", c.MinSecretLength)
	}
	
	// Warn if refresh interval is too close to TTL
	if c.RefreshInterval >= c.DefaultTTL {
		return fmt.Errorf("refresh interval (%v) should be less than default TTL (%v)", 
			c.RefreshInterval, c.DefaultTTL)
	}
	
	return nil
}

// String returns a string representation of the config (without sensitive data)
func (c *SecretConfig) String() string {
	return fmt.Sprintf("SecretConfig{DefaultTTL:%v, RefreshInterval:%v, MinSecretLength:%d, ValidateOnLoad:%t}",
		c.DefaultTTL, c.RefreshInterval, c.MinSecretLength, c.ValidateOnLoad)
}