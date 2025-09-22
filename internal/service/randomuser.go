package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

const (
	randomUserURL = "https://randomuser.me/api/?results=5000"
)

// RandomUser represents a random user from RandomUser API.
type RandomUser struct {
	Results []Results `json:"results"`
	Info    Info      `json:"info"`
}

// Name represents a name of a random user.
type Name struct {
	Title string `json:"title"`
	First string `json:"first"`
	Last  string `json:"last"`
}

// Login represents a login of a random user.
type Login struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Password string `json:"password"`
	Salt     string `json:"salt"`
	Md5      string `json:"md5"`
	Sha1     string `json:"sha1"`
	Sha256   string `json:"sha256"`
}

// Dob represents a date of birth of a random user.
type Dob struct {
	Date time.Time `json:"date"`
	Age  int       `json:"age"`
}

// Registered represents a registered date of a random user.
type Registered struct {
	Date time.Time `json:"date"`
	Age  int       `json:"age"`
}

// ID represents an ID of a random user.
type ID struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Picture represents a picture of a random user.
type Picture struct {
	Large     string `json:"large"`
	Medium    string `json:"medium"`
	Thumbnail string `json:"thumbnail"`
}

// Results represents a result of a random user.
type Results struct {
	Gender     string     `json:"gender"`
	Name       Name       `json:"name"`
	Email      string     `json:"email"`
	Login      Login      `json:"login"`
	Dob        Dob        `json:"dob"`
	Registered Registered `json:"registered"`
	Phone      string     `json:"phone"`
	Cell       string     `json:"cell"`
	ID         ID         `json:"id"`
	Picture    Picture    `json:"picture"`
	Nat        string     `json:"nat"`
}

// Info represents an info of a random user.
type Info struct {
	Seed    string `json:"seed"`
	Results int    `json:"results"`
	Page    int    `json:"page"`
	Version string `json:"version"`
}

// FetchRandomUsers fetches random users from RandomUser API.
func FetchRandomUsers(ctx context.Context, client http.Client) (*RandomUser, error) {
	var out RandomUser
	if err := fetch(ctx, client, randomUserURL, &out); err != nil {
		return nil, errors.Join(ErrRandomUserAPI, err)
	}
	return &out, nil
}

func fetch(ctx context.Context, c http.Client, url string, r any) error {
	const (
		maxRetries = 3
		baseDelay  = 1 * time.Second
		maxDelay   = 30 * time.Second
	)

	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		if err != nil {
			return fmt.Errorf("failed to create GET request: %w", err)
		}

		resp, err := c.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to GET response: %w", err)
			if attempt < maxRetries {
				delay := calculateBackoffDelay(attempt, baseDelay, maxDelay)
				select {
				case <-ctx.Done():
					return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
				case <-time.After(delay):
					continue
				}
			}
			break
		}

		defer resp.Body.Close()

		// Handle different HTTP status codes
		switch resp.StatusCode {
		case http.StatusOK:
			if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}
			return nil

		case http.StatusTooManyRequests:
			lastErr = fmt.Errorf("rate limited (429), attempt %d/%d", attempt+1, maxRetries+1)

			// Check for Retry-After header
			retryAfter := resp.Header.Get("Retry-After")
			var delay time.Duration
			if retryAfter != "" {
				if seconds, parseErr := time.ParseDuration(retryAfter + "s"); parseErr == nil {
					delay = seconds
				} else {
					delay = calculateBackoffDelay(attempt, baseDelay, maxDelay)
				}
			} else {
				delay = calculateBackoffDelay(attempt, baseDelay, maxDelay)
			}

			// Don't retry on the last attempt
			if attempt < maxRetries {
				select {
				case <-ctx.Done():
					return fmt.Errorf("context cancelled during rate limit backoff: %w", ctx.Err())
				case <-time.After(delay):
					continue
				}
			}

		default:
			return fmt.Errorf("unexpected HTTP status: %d", resp.StatusCode)
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries+1, lastErr)
}

// calculateBackoffDelay implements exponential backoff with jitter
func calculateBackoffDelay(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	// Exponential backoff: baseDelay * 2^attempt
	delay := baseDelay * time.Duration(1<<uint(attempt))

	// Cap the delay at maxDelay
	if delay > maxDelay {
		delay = maxDelay
	}

	// Add jitter (randomness) to prevent thundering herd
	// Use full jitter: random delay between 0 and the calculated delay
	// This is simpler and safer than ± jitter
	jitteredDelay := time.Duration(rand.Float64() * float64(delay))

	// Ensure minimum delay is at least baseDelay/4 to avoid too short delays
	minDelay := baseDelay / 4
	if jitteredDelay < minDelay {
		jitteredDelay = minDelay
	}

	return jitteredDelay
}
