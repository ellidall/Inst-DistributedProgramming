package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	commonevent "product/pkg/common/event"
	"product/pkg/product/domain/model"
)

type ItemRequest struct {
	ProductID uuid.UUID
	Count     int
}

type ItemPrice struct {
	ProductID uuid.UUID
	Price     float64
}

type ProductService interface {
	CreateProduct(name string, price float64, stock int) (*model.Product, error)
	UpdateProduct(id uuid.UUID, name string, price float64, stock int) error
	RemoveProduct(id uuid.UUID) error
	ReserveStock(orderID uuid.UUID, items []ItemRequest) ([]ItemPrice, error)
	CancelReservation(orderID uuid.UUID) error
}

func NewProductService(
	productRepo model.ProductRepository,
	reservationRepo model.ReservationRepository,
	dispatcher commonevent.Dispatcher,
) ProductService {
	return &productService{
		productRepo:     productRepo,
		reservationRepo: reservationRepo,
		dispatcher:      dispatcher,
	}
}

type productService struct {
	productRepo     model.ProductRepository
	reservationRepo model.ReservationRepository
	dispatcher      commonevent.Dispatcher
}

func (s *productService) CreateProduct(name string, price float64, stock int) (*model.Product, error) {
	id, err := s.productRepo.NextID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	product := &model.Product{
		ID:        id,
		Name:      name,
		Price:     price,
		Stock:     stock,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.productRepo.Store(product); err != nil {
		return nil, err
	}

	err = s.dispatcher.Dispatch(model.ProductCreated{
		ProductID: id,
		Name:      name,
		Price:     price,
		Stock:     stock,
	})
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (s *productService) UpdateProduct(id uuid.UUID, name string, price float64, stock int) error {
	product, err := s.productRepo.Find(id)
	if err != nil {
		return err
	}

	product.Name = name
	product.Price = price
	product.Stock = stock
	product.UpdatedAt = time.Now()

	if err := s.productRepo.Store(product); err != nil {
		return err
	}

	return s.dispatcher.Dispatch(model.ProductUpdated{
		ProductID: id,
		Name:      name,
		Price:     price,
		Stock:     stock,
	})
}

func (s *productService) RemoveProduct(id uuid.UUID) error {
	product, err := s.productRepo.Find(id)
	if err != nil {
		if errors.Is(err, model.ErrProductNotFound) {
			return nil
		}
		return err
	}

	now := time.Now()
	product.DeletedAt = &now
	product.UpdatedAt = now

	if err := s.productRepo.Store(product); err != nil {
		return err
	}

	return s.dispatcher.Dispatch(model.ProductRemoved{
		ProductID: id,
	})
}

func (s *productService) ReserveStock(orderID uuid.UUID, items []ItemRequest) ([]ItemPrice, error) {
	existingReservations, err := s.reservationRepo.FindKeysByOrderID(orderID)
	if err != nil {
		return nil, err
	}
	if len(existingReservations) > 0 {
		prices := make([]ItemPrice, len(existingReservations))
		for i, r := range existingReservations {
			prices[i] = ItemPrice{
				ProductID: r.ProductID,
				Price:     r.Price,
			}
		}
		return prices, nil
	}

	productIDs := make([]uuid.UUID, len(items))
	for i, item := range items {
		productIDs[i] = item.ProductID
	}

	products, err := s.productRepo.FindMany(productIDs)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		product, exists := products[item.ProductID]
		if !exists {
			return nil, model.ErrProductNotFound
		}
		if product.Stock < item.Count {
			return nil, model.ErrInsufficientStock
		}
	}

	prices := make([]ItemPrice, len(items))
	now := time.Now()

	for i, item := range items {
		product := products[item.ProductID]
		product.Stock -= item.Count
		product.UpdatedAt = now

		if err := s.productRepo.Store(product); err != nil {
			return nil, err
		}

		reservation := &model.ProductReservation{
			OrderID:   orderID,
			ProductID: product.ID,
			Count:     item.Count,
			Price:     product.Price,
			CreatedAt: now,
		}
		if err := s.reservationRepo.Store(reservation); err != nil {
			return nil, err
		}

		prices[i] = ItemPrice{
			ProductID: product.ID,
			Price:     product.Price,
		}
	}

	return prices, nil
}

func (s *productService) CancelReservation(orderID uuid.UUID) error {
	reservations, err := s.reservationRepo.FindKeysByOrderID(orderID)
	if err != nil {
		return err
	}
	if len(reservations) == 0 {
		return nil
	}

	productIDs := make([]uuid.UUID, len(reservations))
	for i, r := range reservations {
		productIDs[i] = r.ProductID
	}

	products, err := s.productRepo.FindMany(productIDs)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, r := range reservations {
		if product, exists := products[r.ProductID]; exists {
			product.Stock += r.Count
			product.UpdatedAt = now
			if err := s.productRepo.Store(product); err != nil {
				return err
			}
		}
	}

	return s.reservationRepo.RemoveByOrderID(orderID)
}
