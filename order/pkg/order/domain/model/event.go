package model

import "github.com/google/uuid"

type OrderItemDTO struct {
	ProductID uuid.UUID `json:"product_id"`
	Count     int       `json:"count"`
}

type OrderCreated struct {
	OrderID    uuid.UUID      `json:"order_id"`
	CustomerID uuid.UUID      `json:"customer_id"`
	Items      []OrderItemDTO `json:"items"`
}

func (e OrderCreated) Type() string {
	return "OrderCreated"
}

type OrderPaid struct {
	OrderID     uuid.UUID `json:"order_id"`
	CustomerID  uuid.UUID `json:"customer_id"`
	TotalAmount float64   `json:"total_amount"`
}

func (e OrderPaid) Type() string {
	return "OrderPaid"
}
