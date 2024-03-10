package service

import (
	"context"

	"wonderful/internal/entities"
	"wonderful/internal/repository"
)

// UserService is a domain service for users.
type UserService interface {
	ListUsers(ctx context.Context, p repository.Params) ([]entities.User, error)
	Create(ctx context.Context) error
}
