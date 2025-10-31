package model

import "github.com/google/uuid"

type OrderItem struct {
	OrderID    uuid.UUID
	ProductID  uuid.UUID
	Count      int
	TotalPrice float64
}

type OrderItemRepository interface {
	Store(orderItem *OrderItem) error
	Find(id uuid.UUID) (*OrderItem, error)
	Remove(id uuid.UUID) error
}
