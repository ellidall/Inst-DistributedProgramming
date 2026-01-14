package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"user/pkg/common/domain"
	"user/pkg/user/domain/model"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) NextID() (uuid.UUID, error) {
	args := m.Called()
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockUserRepository) Store(user model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Find(spec model.FindSpec) (*model.User, error) {
	args := m.Called(spec)
	if user, ok := args.Get(0).(*model.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) HardDelete(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

type MockEventDispatcher struct {
	mock.Mock
}

func (m *MockEventDispatcher) Dispatch(event domain.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func TestCreateUser_Success(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)

	login := "testuser"
	email := "test@example.com"
	telegram := "@test"
	status := model.Active
	userID := uuid.New()

	// Проверка уникальности login, email, telegram — все возвращают ErrUserNotFound
	repo.On("Find", mock.MatchedBy(func(spec model.FindSpec) bool {
		return spec.Login != nil && *spec.Login == login
	})).Return(nil, model.ErrUserNotFound).Once()

	repo.On("Find", mock.MatchedBy(func(spec model.FindSpec) bool {
		return spec.Email != nil && *spec.Email == email
	})).Return(nil, model.ErrUserNotFound).Once()

	repo.On("Find", mock.MatchedBy(func(spec model.FindSpec) bool {
		return spec.Telegram != nil && *spec.Telegram == telegram
	})).Return(nil, model.ErrUserNotFound).Once()

	repo.On("NextID").Return(userID, nil).Once()

	// Проверяем, что Store вызывается с правильными данными
	repo.On("Store", mock.MatchedBy(func(u model.User) bool {
		return u.UserID == userID &&
			u.Login == login &&
			u.Email != nil && *u.Email == email &&
			u.Telegram != nil && *u.Telegram == telegram &&
			u.Status == status &&
			!u.CreatedAt.IsZero() && !u.UpdatedAt.IsZero()
	})).Return(nil).Once()

	// Проверяем событие
	dispatcher.On("Dispatch", mock.MatchedBy(func(e domain.Event) bool {
		created, ok := e.(*model.UserCreated)
		return ok &&
			created.UserID == userID &&
			created.Login == login &&
			created.Email != nil && *created.Email == email &&
			created.Telegram != nil && *created.Telegram == telegram &&
			created.Status == status
	})).Return(nil).Once()

	svc := NewUserService(repo, dispatcher)
	id, err := svc.CreateUser(login, &email, &telegram, status)

	assert.NoError(t, err)
	assert.Equal(t, userID, id)
	repo.AssertExpectations(t)
	dispatcher.AssertExpectations(t)
}

func TestCreateUser_WithNilFields(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)

	login := "testuser"
	var email *string
	var telegram *string
	status := model.Blocked
	userID := uuid.New()

	// Только проверка login
	repo.On("Find", mock.MatchedBy(func(spec model.FindSpec) bool {
		return spec.Login != nil && *spec.Login == login
	})).Return(nil, model.ErrUserNotFound).Once()

	// Email и Telegram не проверяются, так как nil
	repo.On("NextID").Return(userID, nil).Once()

	repo.On("Store", mock.MatchedBy(func(u model.User) bool {
		return u.UserID == userID &&
			u.Login == login &&
			u.Email == nil &&
			u.Telegram == nil &&
			u.Status == status
	})).Return(nil).Once()

	dispatcher.On("Dispatch", mock.MatchedBy(func(e domain.Event) bool {
		created, ok := e.(*model.UserCreated)
		return ok &&
			created.UserID == userID &&
			created.Login == login &&
			created.Email == nil &&
			created.Telegram == nil &&
			created.Status == status
	})).Return(nil).Once()

	svc := NewUserService(repo, dispatcher)
	id, err := svc.CreateUser(login, email, telegram, status)

	assert.NoError(t, err)
	assert.Equal(t, userID, id)
	repo.AssertExpectations(t)
	dispatcher.AssertExpectations(t)
}

func TestCreateUser_LoginExists(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)

	login := "existinguser"
	email := "new@example.com"
	existingUser := &model.User{Login: login}

	repo.On("Find", mock.MatchedBy(func(spec model.FindSpec) bool {
		return spec.Login != nil && *spec.Login == login
	})).Return(existingUser, nil).Once()

	svc := NewUserService(repo, dispatcher)
	_, err := svc.CreateUser(login, &email, nil, model.Active)

	assert.ErrorIs(t, err, model.ErrUserLoginAlreadyUsed)
	repo.AssertExpectations(t)
	// Dispatcher не должен быть вызван
	dispatcher.AssertNotCalled(t, "Dispatch")
}

func TestCreateUser_EmailExists(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)

	login := "newuser"
	email := "existing@example.com"
	existingUser := &model.User{Email: &email}

	// Login свободен
	repo.On("Find", mock.MatchedBy(func(spec model.FindSpec) bool {
		return spec.Login != nil && *spec.Login == login
	})).Return(nil, model.ErrUserNotFound).Once()

	// Email занят
	repo.On("Find", mock.MatchedBy(func(spec model.FindSpec) bool {
		return spec.Email != nil && *spec.Email == email
	})).Return(existingUser, nil).Once()

	svc := NewUserService(repo, dispatcher)
	_, err := svc.CreateUser(login, &email, nil, model.Active)

	assert.ErrorIs(t, err, model.ErrUserEmailAlreadyUsed)
	repo.AssertExpectations(t)
	dispatcher.AssertNotCalled(t, "Dispatch")
}

func TestCreateUser_TelegramExists(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)

	login := "newuser"
	telegram := "@existing"
	existingUser := &model.User{Telegram: &telegram}

	// Login свободен
	repo.On("Find", mock.MatchedBy(func(spec model.FindSpec) bool {
		return spec.Login != nil && *spec.Login == login
	})).Return(nil, model.ErrUserNotFound).Once()

	// Telegram занят
	repo.On("Find", mock.MatchedBy(func(spec model.FindSpec) bool {
		return spec.Telegram != nil && *spec.Telegram == telegram
	})).Return(existingUser, nil).Once()

	svc := NewUserService(repo, dispatcher)
	_, err := svc.CreateUser(login, nil, &telegram, model.Active)

	assert.ErrorIs(t, err, model.ErrUserTelegramAlreadyUsed)
	repo.AssertExpectations(t)
	dispatcher.AssertNotCalled(t, "Dispatch")
}

func TestUpdateUser_Success(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)

	userID := uuid.New()
	existingUser := &model.User{UserID: userID, Login: "test"}

	repo.On("Find", mock.Anything).Return(existingUser, nil)
	repo.On("Store", mock.Anything).Return(nil)
	dispatcher.On("Dispatch", mock.AnythingOfType("*model.UserUpdated")).Return(nil)

	svc := NewUserService(repo, dispatcher)
	err := svc.UpdateUser(userID, struct {
		Status   *model.UserStatus
		Email    *string
		Telegram *string
	}{
		Status: func() *model.UserStatus { s := model.Active; return &s }(),
	})

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	dispatcher.AssertExpectations(t)
}

func TestDeleteUser_Success(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)

	userID := uuid.New()
	existingUser := &model.User{UserID: userID}

	repo.On("Find", mock.Anything).Return(existingUser, nil)
	repo.On("Store", mock.Anything).Return(nil)
	dispatcher.On("Dispatch", mock.AnythingOfType("*model.UserDeleted")).Return(nil)

	svc := NewUserService(repo, dispatcher)
	err := svc.DeleteUser(userID, false)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	dispatcher.AssertExpectations(t)
}
