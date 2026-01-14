package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"notification/pkg/common/domain"
	"notification/pkg/notification/domain/model"
)

const email = "email"

type UpdateUserParams struct {
	Status   *model.UserStatus
	Email    *string
	Telegram *string
}

type UserService interface {
	CreateUser(
		userID uuid.UUID,
		login string,
		email *string,
		telegram *string,
		status model.UserStatus,
	) error
	UpdateUser(userID uuid.UUID, params UpdateUserParams) error
	DeleteUser(userID uuid.UUID, hard bool) error
}

func NewUserService(
	userRepository model.UserRepository,
	eventDispatcher domain.EventDispatcher,
) UserService {
	return &userService{
		userRepository:  userRepository,
		eventDispatcher: eventDispatcher,
	}
}

type userService struct {
	userRepository  model.UserRepository
	eventDispatcher domain.EventDispatcher
}

func (u userService) CreateUser(
	userID uuid.UUID,
	login string,
	email *string,
	telegram *string,
	status model.UserStatus,
) error {
	fmt.Println("CreateUser")
	// Проверка уникальности login
	_, err := u.userRepository.Find(model.FindSpec{Login: &login})
	if err == nil {
		return model.ErrUserLoginAlreadyUsed
	}
	if !errors.Is(err, model.ErrUserNotFound) {
		return err
	}

	// Проверка уникальности email
	if email != nil && *email != "" {
		_, err := u.userRepository.Find(model.FindSpec{Email: email})
		if err == nil {
			return model.ErrUserEmailAlreadyUsed
		}
		if !errors.Is(err, model.ErrUserNotFound) {
			return err
		}
	}

	// Проверка уникальности telegram
	if telegram != nil && *telegram != "" {
		_, err := u.userRepository.Find(model.FindSpec{Telegram: telegram})
		if err == nil {
			return model.ErrUserTelegramAlreadyUsed
		}
		if !errors.Is(err, model.ErrUserNotFound) {
			return err
		}
	}

	currentTime := time.Now()
	user := model.User{
		UserID:    userID,
		Status:    status,
		Login:     login,
		Email:     email,
		Telegram:  telegram,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	return u.userRepository.Store(user)
}

func (u userService) UpdateUser(userID uuid.UUID, params UpdateUserParams) error {
	user, err := u.userRepository.Find(model.FindSpec{UserID: &userID})
	if err != nil {
		return err
	}

	currentTime := time.Now()
	hasChanges := false

	// Статус
	if params.Status != nil && user.Status != *params.Status {
		user.Status = *params.Status
		hasChanges = true
	}

	// Email
	if changed, err := u.handleFieldUpdate(user, "email", params.Email); err != nil {
		return err
	} else if changed {
		user.Email = params.Email
		hasChanges = true
	}

	// Telegram
	if changed, err := u.handleFieldUpdate(user, "telegram", params.Telegram); err != nil {
		return err
	} else if changed {
		user.Telegram = params.Telegram
		hasChanges = true
	}

	if !hasChanges {
		return nil
	}

	user.UpdatedAt = currentTime
	if err := u.userRepository.Store(*user); err != nil {
		return err
	}

	// Публикуем событие ОДИН РАЗ
	event := model.UserUpdated{
		UserID:    userID,
		UpdatedAt: currentTime.UnixMilli(),
	}

	updatedFields := &model.UpdatedFields{}
	hasUpdated := false

	if params.Status != nil {
		updatedFields.Status = params.Status
		hasUpdated = true
	}
	if params.Email != nil {
		updatedFields.Email = params.Email
		hasUpdated = true
	}
	if params.Telegram != nil {
		updatedFields.Telegram = params.Telegram
		hasUpdated = true
	}

	if hasUpdated {
		event.UpdatedFields = updatedFields
	}

	// Обработка удаления полей
	if params.Email == nil && user.Email == nil {
		// Поле было удалено
		if event.RemovedFields == nil {
			event.RemovedFields = &model.RemovedFields{}
		}
		event.RemovedFields.Email = toPtr(true)
	}
	if params.Telegram == nil && user.Telegram == nil {
		if event.RemovedFields == nil {
			event.RemovedFields = &model.RemovedFields{}
		}
		event.RemovedFields.Telegram = toPtr(true)
	}

	return nil
}

func (u userService) handleFieldUpdate(user *model.User, field string, newValue *string) (bool, error) {
	var currentValue *string
	switch field {
	case email:
		currentValue = user.Email
	case "telegram":
		currentValue = user.Telegram
	}

	// Удаление поля
	if newValue == nil {
		return currentValue != nil, nil
	}

	// Проверка на пустое значение
	if *newValue == "" {
		if currentValue != nil {
			return true, nil // Удаляем существующее значение
		}
		return false, nil // Нечего удалять
	}

	// Проверка на изменение значения
	valuesEqual := currentValue != nil && *currentValue == *newValue
	if valuesEqual {
		return false, nil
	}

	// Проверка уникальности
	spec := model.FindSpec{}
	if field == email {
		spec.Email = newValue
	} else {
		spec.Telegram = newValue
	}

	if existing, _ := u.userRepository.Find(spec); existing != nil && existing.UserID != user.UserID {
		if field == email {
			return false, model.ErrUserEmailAlreadyUsed
		}
		return false, model.ErrUserTelegramAlreadyUsed
	}

	return true, nil
}

func (u userService) DeleteUser(userID uuid.UUID, hard bool) error {
	user, err := u.userRepository.Find(model.FindSpec{UserID: &userID})
	if err != nil {
		return err
	}

	if hard {
		err = u.userRepository.HardDelete(userID)
		if err != nil {
			return err
		}
	}

	currentTime := time.Now()
	user.Status = model.Deleted
	user.UpdatedAt = currentTime
	user.DeletedAt = &currentTime

	return u.userRepository.Store(*user)
}

func toPtr[T any](v T) *T {
	return &v
}
