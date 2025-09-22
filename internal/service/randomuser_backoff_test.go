package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateBackoffDelay(t *testing.T) {
	baseDelay := 1 * time.Second
	maxDelay := 30 * time.Second

	tests := []struct {
		name           string
		attempt        int
		expectedBase   time.Duration
		maxExpectedMax time.Duration
	}{
		{
			name:           "first attempt",
			attempt:        0,
			expectedBase:   1 * time.Second, // 1 * 2^0 = 1s
			maxExpectedMax: 2 * time.Second, // with jitter should be less than 2x
		},
		{
			name:           "second attempt",
			attempt:        1,
			expectedBase:   2 * time.Second, // 1 * 2^1 = 2s
			maxExpectedMax: 3 * time.Second, // with jitter should be reasonable
		},
		{
			name:           "third attempt",
			attempt:        2,
			expectedBase:   4 * time.Second, // 1 * 2^2 = 4s
			maxExpectedMax: 6 * time.Second, // with jitter should be reasonable
		},
		{
			name:           "max delay reached",
			attempt:        10,
			expectedBase:   maxDelay,         // Should be capped at 30s
			maxExpectedMax: 40 * time.Second, // with jitter, slightly higher is OK
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run multiple times to test jitter variability
			for i := 0; i < 5; i++ {
				delay := calculateBackoffDelay(tt.attempt, baseDelay, maxDelay)

				// Check that delay is always positive
				assert.Positive(t, delay, "delay should always be positive")

				// Check that delay is not too large
				assert.LessOrEqual(t, delay, tt.maxExpectedMax, "delay should not exceed reasonable bounds")

				// Check that delay is at least some minimum (not too small with negative jitter)
				minDelay := tt.expectedBase / 4 // Allow for significant jitter
				assert.GreaterOrEqual(t, delay, minDelay, "delay should not be too small")
			}
		})
	}
}

func TestCalculateBackoffDelayBounds(t *testing.T) {
	baseDelay := 1 * time.Second
	maxDelay := 5 * time.Second

	// Test that very high attempts are still reasonable
	delay := calculateBackoffDelay(100, baseDelay, maxDelay)

	assert.Positive(t, delay)
	assert.LessOrEqual(t, delay, 10*time.Second) // Should not be way higher than maxDelay
}
