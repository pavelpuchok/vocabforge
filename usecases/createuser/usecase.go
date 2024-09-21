package createuser

import (
	"context"
	"fmt"

	"github.com/pavelpuchok/vocabforge/models"
)

type UseCase struct {
	UsersService UsersService
}

type UsersService interface {
	Create(ctx context.Context) (models.User, error)
}

func (u UseCase) Run(ctx context.Context) (models.User, error) {
	usr, err := u.UsersService.Create(ctx)
	if err != nil {
		return usr, fmt.Errorf("create_user.UseCase.Run unable to create user. %w", err)
	}
	return usr, nil
}
