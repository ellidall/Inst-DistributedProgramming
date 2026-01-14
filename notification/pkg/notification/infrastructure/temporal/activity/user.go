package activity

import (
	"context"

	"github.com/google/uuid"

	appdata "notification/pkg/notification/app/data"
	"notification/pkg/notification/app/service"
)

func NewUserActivities(userService service.UserService) *UserActivities {
	return &UserActivities{userService: userService}
}

type UserActivities struct {
	userService service.UserService
}

func (a *UserActivities) FindUser(ctx context.Context, userID uuid.UUID) (appdata.User, error) {
	return a.userService.FindUser(ctx, userID)
}

func (a *UserActivities) UpdateUser(ctx context.Context, userID uuid.UUID, update appdata.UserUpdate) error {
	return a.userService.UpdateUser(ctx, userID, update)
}

func (a *UserActivities) CreateUser(ctx context.Context, user appdata.User) error {
	return a.userService.CreateUser(ctx, user)
}
