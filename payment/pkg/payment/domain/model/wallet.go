package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrWalletNotFound    = errors.New("wallet not found")
	ErrInsufficientFunds = errors.New("insufficient funds") // Новая ошибка
)

type Wallet struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Balance   float64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type WalletRepository interface {
	NextID() (uuid.UUID, error)
	Store(wallet *Wallet) error
	Find(id uuid.UUID) (*Wallet, error)
	FindByUserID(userID uuid.UUID) (*Wallet, error)
	Remove(id uuid.UUID) error
}
