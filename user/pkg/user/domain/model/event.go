package model

import (
	"github.com/google/uuid"
)

type UserCreated struct {
	UserID    uuid.UUID  `json:"user_id"`
	Status    UserStatus `json:"status"`
	Login     string     `json:"login"`
	Email     *string    `json:"email,omitempty"`
	Telegram  *string    `json:"telegram,omitempty"`
	CreatedAt int64      `json:"created_at"`
}

func (u UserCreated) Type() string {
	return "user_created"
}

type UpdatedFields struct {
	Status   *UserStatus `json:"status,omitempty"`
	Email    *string     `json:"email,omitempty"`
	Telegram *string     `json:"telegram,omitempty"`
}

type RemovedFields struct {
	Email    *bool `json:"email,omitempty"`
	Telegram *bool `json:"telegram,omitempty"`
}

type UserUpdated struct {
	UserID        uuid.UUID      `json:"user_id"`
	UpdatedFields *UpdatedFields `json:"updated_fields,omitempty"`
	RemovedFields *RemovedFields `json:"removed_fields,omitempty"`
	UpdatedAt     int64          `json:"updated_at"`
}

func (u UserUpdated) Type() string {
	return "user_updated"
}

type UserDeleted struct {
	UserID    uuid.UUID  `json:"user_id"`
	Status    UserStatus `json:"status"`
	DeletedAt int64      `json:"deleted_at"`
	Hard      bool       `json:"hard"`
}

func (u UserDeleted) Type() string {
	return "user_deleted"
}
