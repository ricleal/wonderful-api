// Package testhelpers contains assorted utilities to assist with writing tests
// which use HTTP.
package testhelpers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	contentTypeKey   = "Content-Type"
	contentTypeValue = "application/json" // Changed from application/vnd.api+json
	// Default JWT secret for testing
	TestJWTSecret = "production-grade-jwt-signing-key-for-automated-testing-purposes"
)

// JWTClaims represents the JWT claims structure for testing
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateTestJWT generates a JWT token for testing purposes
func GenerateTestJWT(userID, email string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "wonderful-api",
			Subject:   userID,
			ID:        fmt.Sprintf("jwt-test-%d", now.Unix()),
			Audience:  []string{"wonderful-api"},
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token
	tokenString, err := token.SignedString([]byte(TestJWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT token: %w", err)
	}

	return tokenString, nil
}

// GetJWTHeaders returns headers with a valid JWT token for testing
func GetJWTHeaders(userID, email string) (map[string]string, error) {
	token, err := GenerateTestJWT(userID, email, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"Authorization": "Bearer " + token,
	}, nil
}

func parseResponse(req *http.Request, response interface{}) (int, error) {
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}
	// if there's no reply body and we not expect one, we're done
	if len(body) == 0 {
		if responseStruct, ok := response.(*struct{}); ok && *responseStruct == struct{}{} {
			return res.StatusCode, nil
		}
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	return res.StatusCode, nil
}

// Get executes an HTTP GET request to the given URL. The response body is
// assumed to contain JSON and this is unmarshaled into the provided response
// object. The HTTP response status code is also returned.
func Get(ctx context.Context, theURL string, response interface{}) (statusCode int, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, theURL, http.NoBody)
	if err != nil {
		return 0, fmt.Errorf("failed to GET request: %w", err)
	}
	return parseResponse(req, response)
}

// GetWithHeaders is the same as Get but it also provides the given HTTP headers
// in the request.
func GetWithHeaders(
	ctx context.Context,
	theURL string,
	headers map[string]string,
	response interface{},
) (statusCode int, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, theURL, http.NoBody)
	if err != nil {
		return 0, fmt.Errorf("failed to GET request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return parseResponse(req, response)
}

// Post executes an HTTP POST request to the given URL with given body.
// The content type is set to application/vnd.api+json. The response body is
// assumed to contain JSON and this is unmarshaled into the provided response
// object. The HTTP response status code is also returned.
func Post(ctx context.Context, theURL, body string, response interface{}) (statusCode int, err error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, theURL, strings.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("failed to POST request: %w", err)
	}
	req.Header.Set(contentTypeKey, contentTypeValue)
	return parseResponse(req, response)
}

// PostWithHeaders is the same as Post but it also provides the given HTTP headers
// in the request.
func PostWithHeaders(
	ctx context.Context,
	theURL string,
	headers map[string]string,
	body string,
	response interface{},
) (statusCode int, err error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, theURL, strings.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("failed to POST request: %w", err)
	}
	req.Header.Set(contentTypeKey, contentTypeValue)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return parseResponse(req, response)
}

// Patch executes an HTTP PATCH request to the given URL with given body.
func Patch(ctx context.Context, theURL, body string, headers map[string]string, response interface{}) (statusCode int, err error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPatch, theURL, strings.NewReader(body),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to PATCH request: %w", err)
	}
	req.Header.Set(contentTypeKey, contentTypeValue)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return parseResponse(req, response)
}

// Delete executes an HTTP DELETE request to the given URL.
func Delete(ctx context.Context, theURL string, response interface{}) (statusCode int, err error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodDelete, theURL, http.NoBody)
	if err != nil {
		return 0, fmt.Errorf("failed to DELETE request: %w", err)
	}
	return parseResponse(req, response)
}

// DeleteWithHeaders is the same as Delete but it also provides the given HTTP headers.
func DeleteWithHeaders(ctx context.Context, theURL string, headers map[string]string, response interface{}) (statusCode int, err error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodDelete, theURL, http.NoBody)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to DELETE request: %w", err)
	}
	return parseResponse(req, response)
}

// GetWithJWT executes an HTTP GET request with JWT authentication
func GetWithJWT(ctx context.Context, theURL, userID, email string, response interface{}) (statusCode int, err error) {
	headers, err := GetJWTHeaders(userID, email)
	if err != nil {
		return 0, fmt.Errorf("failed to generate JWT headers: %w", err)
	}
	return GetWithHeaders(ctx, theURL, headers, response)
}

// PostWithJWT executes an HTTP POST request with JWT authentication
func PostWithJWT(ctx context.Context, theURL, userID, email, body string, response interface{}) (statusCode int, err error) {
	headers, err := GetJWTHeaders(userID, email)
	if err != nil {
		return 0, fmt.Errorf("failed to generate JWT headers: %w", err)
	}
	return PostWithHeaders(ctx, theURL, headers, body, response)
}
