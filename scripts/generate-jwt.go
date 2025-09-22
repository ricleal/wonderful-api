package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the JWT claims structure (same as in middleware)
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func main() {
	var (
		userID   = flag.String("user-id", "test-user-123", "User ID for the JWT token")
		email    = flag.String("email", "test@example.com", "Email for the JWT token")
		duration = flag.Duration("duration", 24*time.Hour, "Token expiration duration")
		secret   = flag.String("secret", "", "JWT secret (or use JWT_SECRET env var)")
	)
	flag.Parse()

	// Get JWT secret from flag or environment
	jwtSecret := *secret
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			log.Fatal("JWT secret is required. Use -secret flag or set JWT_SECRET environment variable.")
		}
	}

	if len(jwtSecret) < 32 {
		log.Fatal("JWT secret must be at least 32 characters long for security.")
	}

	// Create the Claims
	now := time.Now()
	claims := JWTClaims{
		UserID: *userID,
		Email:  *email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(*duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "wonderful-api",
			Subject:   *userID,
			ID:        fmt.Sprintf("jwt-%d", now.Unix()),
			Audience:  []string{"wonderful-api"},
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Fatalf("Error generating token: %v", err)
	}

	// Print just the token
	fmt.Print(tokenString)
}
