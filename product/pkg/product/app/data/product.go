package data

import "github.com/google/uuid"

type Product struct {
	ID    uuid.UUID
	Name  string
	Price float64
	Stock int
}

type ReserveItem struct {
	ProductID uuid.UUID
	Count     int
}

type ItemPrice struct {
	ProductID uuid.UUID
	Price     float64
}
