package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// JWTContextKey is the context key for JWT claims
type JWTContextKey string

const (
	ClaimsContextKey JWTContextKey = "jwt_claims"
)

// JWTAuth middleware validates JWT tokens
func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get JWT secret from environment
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"code": 500, "message": "JWT secret not configured"}`, http.StatusInternalServerError)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"code": 401, "message": "Authorization header required"}`, http.StatusUnauthorized)
			return
		}

		// Check for Bearer token format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"code": 401, "message": "Authorization header must start with Bearer"}`, http.StatusUnauthorized)
			return
		}

		// Extract token string
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"code": 401, "message": "Token is required"}`, http.StatusUnauthorized)
			return
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, fmt.Sprintf(`{"code": 401, "message": "Invalid token: %s"}`, err.Error()), http.StatusUnauthorized)
			return
		}

		// Validate token and extract claims
		if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
			// Add claims to request context
			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"code": 401, "message": "Invalid token claims"}`, http.StatusUnauthorized)
			return
		}
	})
}

// GetJWTClaims extracts JWT claims from request context
func GetJWTClaims(ctx context.Context) (*JWTClaims, error) {
	claims, ok := ctx.Value(ClaimsContextKey).(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("JWT claims not found in context")
	}
	return claims, nil
}
