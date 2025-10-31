package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrOrderNotFound = errors.New("order not found")

type OrderStatus int

const (
	Open OrderStatus = iota
	Pending
	Paid
	Cancelled
)

type Order struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	Status     OrderStatus
	Items      []OrderItem
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

type OrderRepository interface {
	NextID() (uuid.UUID, error)
	Store(order *Order) error
	Find(id uuid.UUID) (*Order, error)
	Remove(id uuid.UUID) error
}
