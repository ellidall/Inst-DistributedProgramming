package service

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/google/uuid"

	"notification/pkg/common/domain"
	appdata "notification/pkg/notification/app/data"
	"notification/pkg/notification/domain/model"
	"notification/pkg/notification/domain/service"
)

type UserService interface {
	CreateUser(ctx context.Context, user appdata.User) error
	UpdateUser(ctx context.Context, userID uuid.UUID, update appdata.UserUpdate) error
	BlockUser(ctx context.Context, userID uuid.UUID) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	FindUser(ctx context.Context, userID uuid.UUID) (appdata.User, error)
}

func NewUserService(
	uow UnitOfWork,
	luow LockableUnitOfWork,
	eventDispatcher outbox.EventDispatcher[outbox.Event],
) UserService {
	return &userService{
		uow:             uow,
		luow:            luow,
		eventDispatcher: eventDispatcher,
	}
}

type userService struct {
	uow             UnitOfWork
	luow            LockableUnitOfWork
	eventDispatcher outbox.EventDispatcher[outbox.Event]
}

func (s *userService) CreateUser(ctx context.Context, user appdata.User) error {
	var lockNames []string
	lockNames = append(lockNames, userLoginLock(user.Login))
	if user.Email != nil && *user.Email != "" {
		lockNames = append(lockNames, userEmailLock(*user.Email))
	}
	if user.Telegram != nil && *user.Telegram != "" {
		lockNames = append(lockNames, userTelegramLock(*user.Telegram))
	}

	return s.luow.Execute(ctx, lockNames, func(provider RepositoryProvider) error {
		status := model.Blocked
		if user.Status != 0 {
			status = model.UserStatus(user.Status)
		}

		domainService := s.domainService(ctx, provider.UserRepository(ctx))
		return domainService.CreateUser(user.UserID, user.Login, user.Email, user.Telegram, status)
	})
}

func (s *userService) UpdateUser(ctx context.Context, userID uuid.UUID, update appdata.UserUpdate) error {
	var lockNames []string
	lockNames = append(lockNames, userLock(userID))
	if update.Email != nil && *update.Email != "" {
		lockNames = append(lockNames, userEmailLock(*update.Email))
	}
	if update.Telegram != nil && *update.Telegram != "" {
		lockNames = append(lockNames, userTelegramLock(*update.Telegram))
	}

	return s.luow.Execute(ctx, lockNames, func(provider RepositoryProvider) error {
		domainService := s.domainService(ctx, provider.UserRepository(ctx))
		params := s.convertToUpdateParams(update)
		return domainService.UpdateUser(userID, params) // ← только публикация события
	})
}

func (s *userService) BlockUser(ctx context.Context, userID uuid.UUID) error {
	return s.luow.Execute(ctx, []string{userLock(userID)}, func(provider RepositoryProvider) error {
		domainService := s.domainService(ctx, provider.UserRepository(ctx))
		status := model.Blocked
		return domainService.UpdateUser(userID, struct {
			Status   *model.UserStatus
			Email    *string
			Telegram *string
		}{Status: &status})
	})
}

func (s *userService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return s.luow.Execute(ctx, []string{userLock(userID)}, func(provider RepositoryProvider) error {
		return s.domainService(ctx, provider.UserRepository(ctx)).DeleteUser(userID, false)
	})
}

func (s *userService) FindUser(ctx context.Context, userID uuid.UUID) (appdata.User, error) {
	var user appdata.User
	err := s.luow.Execute(ctx, []string{userLock(userID)}, func(provider RepositoryProvider) error {
		domainUser, err := provider.UserRepository(ctx).Find(model.FindSpec{UserID: &userID})
		if err != nil {
			return err
		}
		user = appdata.User{
			UserID:   domainUser.UserID,
			Status:   int(domainUser.Status),
			Login:    domainUser.Login,
			Email:    domainUser.Email,
			Telegram: domainUser.Telegram,
		}
		return nil
	})
	return user, err
}

func (s *userService) convertToUpdateParams(update appdata.UserUpdate) service.UpdateUserParams {
	params := service.UpdateUserParams{}
	if update.Status != nil {
		status := model.UserStatus(*update.Status)
		params.Status = &status
	}
	params.Email = update.Email
	params.Telegram = update.Telegram
	return params
}

func (s *userService) domainService(ctx context.Context, repository model.UserRepository) service.UserService {
	return service.NewUserService(repository, s.domainEventDispatcher(ctx))
}

func (s *userService) domainEventDispatcher(ctx context.Context) domain.EventDispatcher {
	return &domainEventDispatcher{
		ctx:             ctx,
		eventDispatcher: s.eventDispatcher,
	}
}

const baseUserLock = "user_"

func userLock(id uuid.UUID) string {
	return baseUserLock + id.String()
}

func userLoginLock(login string) string {
	return baseUserLock + "login_" + login
}

func userEmailLock(email string) string {
	return baseUserLock + "email_" + email
}

func userTelegramLock(telegram string) string {
	return baseUserLock + "telegram_" + telegram
}
