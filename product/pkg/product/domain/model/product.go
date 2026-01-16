package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrReservationExists = errors.New("reservation already exists")
)

type Product struct {
	ID        uuid.UUID
	Name      string
	Price     float64
	Stock     int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type ProductReservation struct {
	OrderID   uuid.UUID
	ProductID uuid.UUID
	Count     int
	Price     float64
	CreatedAt time.Time
}

type ProductRepository interface {
	NextID() (uuid.UUID, error)
	Store(product *Product) error
	Find(id uuid.UUID) (*Product, error)
	FindMany(ids []uuid.UUID) (map[uuid.UUID]*Product, error)
	Remove(id uuid.UUID) error
}

type ReservationRepository interface {
	Store(reservation *ProductReservation) error
	FindKeysByOrderID(orderID uuid.UUID) ([]ProductReservation, error)
	RemoveByOrderID(orderID uuid.UUID) error
}
