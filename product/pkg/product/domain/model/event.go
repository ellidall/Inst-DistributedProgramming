package model

import "github.com/google/uuid"

type ProductCreated struct {
	ProductID uuid.UUID `json:"product_id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Stock     int       `json:"stock"`
}

func (e ProductCreated) Type() string {
	return "ProductCreated"
}

type ProductUpdated struct {
	ProductID uuid.UUID `json:"product_id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Stock     int       `json:"stock"`
}

func (e ProductUpdated) Type() string {
	return "ProductUpdated"
}

type ProductRemoved struct {
	ProductID uuid.UUID `json:"product_id"`
}

func (e ProductRemoved) Type() string {
	return "ProductRemoved"
}
