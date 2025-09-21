#!/bin/bash

# JWT Token Generator Script for Wonderful API
# This script generates JWT tokens for testing the API

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "JWT Token Generator for Wonderful API"
echo "====================================="

# Check if JWT_SECRET is set
if [ -z "$JWT_SECRET" ]; then
    echo "Warning: JWT_SECRET environment variable is not set."
    echo "Please set it or use the -secret flag."
    echo ""
fi

echo "Usage examples:"
echo "  ./scripts/generate-jwt.sh                                    # Generate default token"
echo "  ./scripts/generate-jwt.sh -user-id=\"john123\" -email=\"john@example.com\"  # Custom user"
echo "  ./scripts/generate-jwt.sh -duration=\"1h\"                     # 1 hour expiration"
echo "  ./scripts/generate-jwt.sh -secret=\"your-secret-key\"          # Custom secret"
echo ""

# Run the Go script with all arguments passed through
cd "$PROJECT_ROOT"
go run ./scripts/generate-jwt.go "$@"