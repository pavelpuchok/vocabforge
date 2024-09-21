package users

import (
	"context"
	"fmt"

	"github.com/pavelpuchok/vocabforge/models"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{
		repo,
	}
}

type Repository interface {
	Create(ctx context.Context) (models.User, error)
}

func (s Service) Create(ctx context.Context) (models.User, error) {
	u, err := s.repo.Create(ctx)
	if err != nil {
		return u, fmt.Errorf("users.Service.Create failed. %w", err)
	}
	return u, nil
}
