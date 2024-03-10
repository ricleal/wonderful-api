package service

import (
	"context"
	"fmt"
	"net/http"

	"wonderful/internal/entities"
	"wonderful/internal/repository"
	"wonderful/internal/store"
)

// userService is an implementation of the UserService interface.
type userService struct {
	repo   repository.UserRepository
	client http.Client
}

// NewUserService creates a new UserService.
func NewUserService(s store.Store, c http.Client) *userService {
	return &userService{
		repo:   s.Users(),
		client: c,
	}
}

func (s *userService) Create(ctx context.Context) error {
	rUsers, err := FetchRandomUsers(ctx, s.client)
	if err != nil {
		return fmt.Errorf("service failed to get random users: %w", err)
	}
	// populate a repository with random users.
	repoUsers := make([]repository.User, 0, len(rUsers.Results))
	for i := range rUsers.Results {
		u := rUsers.Results[i] // to avoid creating a new variable in each iteration.
		repoUsers = append(repoUsers, repository.User{
			Name:  u.Name.Title + " " + u.Name.First + " " + u.Name.Last,
			Email: u.Email,
			Phone: u.Phone,
			Cell:  u.Cell,
			Picture: map[string]string{
				"large":     u.Picture.Large,
				"medium":    u.Picture.Medium,
				"thumbnail": u.Picture.Thumbnail,
			},
			Registration: u.Registered.Date,
		})
	}
	// insert random users into the repository.
	if err := s.repo.Create(ctx, repoUsers); err != nil {
		return fmt.Errorf("service failed to insert random users: %w", err)
	}
	return nil
}

func (s *userService) ListUsers(ctx context.Context, p repository.Params) ([]entities.User, error) {
	// fetch users from the repository.
	repoUsers, err := s.repo.ListUsers(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("service failed to list users: %w", err)
	}
	// convert repository users to entities users.
	entitiesUsers := make([]entities.User, 0, len(repoUsers))
	for i := range repoUsers {
		u := repoUsers[i] // to avoid creating a new variable in each iteration.
		entitiesUsers = append(entitiesUsers, entities.User{
			ID:           u.ID.String(),
			Name:         u.Name,
			Email:        u.Email,
			Phone:        u.Phone,
			Cell:         u.Cell,
			Picture:      u.Picture,
			Registration: u.Registration,
		})
	}
	return entitiesUsers, nil
}
