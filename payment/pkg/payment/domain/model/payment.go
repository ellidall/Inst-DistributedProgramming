package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrInvalidPaymentStatus = errors.New("invalid payment status transition")
)

type PaymentStatus int

const (
	Pending PaymentStatus = iota
	Processing
	Succeeded
	Failed
	Cancelled
)

type Payment struct {
	ID        uuid.UUID
	WalletID  uuid.UUID // ID кошелька, с которого платили
	OrderID   uuid.UUID
	Amount    float64
	Status    PaymentStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type PaymentRepository interface {
	NextID() (uuid.UUID, error)
	Store(payment *Payment) error
	Find(id uuid.UUID) (*Payment, error)
	Remove(id uuid.UUID) error
}
