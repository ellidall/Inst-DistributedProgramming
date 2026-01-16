package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrInvalidOrderStatus = errors.New("invalid order status transition")
	ErrEmptyOrderItems    = errors.New("order must have at least one item")
)

type OrderStatus int

const (
	Pending OrderStatus = iota
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

type OrderItem struct {
	OrderID   uuid.UUID
	ProductID uuid.UUID
	Count     int
	Price     float64
}

func (o *Order) CalculateTotal() float64 {
	var total float64
	for _, item := range o.Items {
		total += item.Price * float64(item.Count)
	}
	return total
}

type OrderRepository interface {
	Store(order *Order) error
	Find(id uuid.UUID) (*Order, error)
}
