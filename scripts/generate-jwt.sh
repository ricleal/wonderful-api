#!/bin/bash

# JWT Token Generator Script for Wonderful API
# This script generates JWT tokens for testing the API and exports them as $TOKEN

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Check if we're in export-only mode
if [ "$1" = "--export-only" ]; then
    shift  # Remove the --export-only flag from arguments
    
    # Run the Go script with remaining arguments and capture the token directly
    cd "$PROJECT_ROOT"
    token=$(go run ./scripts/generate-jwt.go "$@")
    
    # Only output the export command
    echo "export TOKEN=\"$token\""
    exit 0
fi

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

# Run the Go script with all arguments passed through and capture the token directly
cd "$PROJECT_ROOT"
token=$(go run ./scripts/generate-jwt.go "$@")

# Print the token with some context
echo "Generated JWT Token:"
echo "Token: $token"
echo ""
echo "Claims: test-user-123, test@example.com (24h expiration)"

# Print instructions for using the exported TOKEN variable
echo ""
echo "⚠️  Note: Running this script directly does NOT export TOKEN to your shell!"
echo "   (Child processes cannot modify parent shell environment)"
echo ""
echo "🎯 To export TOKEN to your current shell session, use one of these:"
echo "   eval \"\$(./scripts/generate-jwt.sh --export-only)\""
echo "   source <(./scripts/generate-jwt.sh --export-only)"
echo ""
echo "💡 For convenience, you can create an alias:"
echo "   alias jwt='eval \"\$(./scripts/generate-jwt.sh --export-only)\"'"
echo "   Then just run: jwt"
echo ""
echo "🧪 Test the exported token with curl:"
echo "   curl -H \"Authorization: Bearer \$TOKEN\" http://localhost:8888/api/v1/wonderfuls"
echo "   curl -H \"Authorization: Bearer \$TOKEN\" http://localhost:8888/api/v1/wonderfuls/SOME_USER_ID"