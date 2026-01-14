package data

import "github.com/google/uuid"

type User struct {
	UserID   uuid.UUID
	Status   int
	Login    string
	Email    *string
	Telegram *string
}

type UserUpdate struct {
	Status   *int
	Email    *string
	Telegram *string
}
