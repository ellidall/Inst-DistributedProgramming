package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"orderservice/pkg/common/event"
	"orderservice/pkg/orderservice/domain/model"
)

var (
	ErrInvalidOrderStatus = errors.New("invalid order status")
)

type OrderService interface {
	CreateOrder(customerID uuid.UUID) (uuid.UUID, error)
	RemoveOrder(orderID uuid.UUID) error
	SetStatus(orderID uuid.UUID, status model.OrderStatus) error

	AddItem(orderID, productID uuid.UUID, price float64) (uuid.UUID, error)
	RemoveItem(orderID, itemID uuid.UUID) error
}

func NewOrderService(
	orderRepo model.OrderRepository,
	orderItemRepo model.OrderItemRepository,
	eventDispatcher event.Dispatcher,
) OrderService {
	return &orderService{
		orderRepo:       orderRepo,
		orderItemRepo:   orderItemRepo,
		eventDispatcher: eventDispatcher,
	}
}

type orderService struct {
	orderRepo       model.OrderRepository
	orderItemRepo   model.OrderItemRepository
	eventDispatcher event.Dispatcher
}

func (o *orderService) CreateOrder(customerID uuid.UUID) (uuid.UUID, error) {
	orderID, err := o.orderRepo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	currentTime := time.Now()
	err = o.orderRepo.Store(&model.Order{
		ID:         orderID,
		CustomerID: customerID,
		Status:     model.Open,
		CreatedAt:  currentTime,
		UpdatedAt:  currentTime,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return orderID, o.eventDispatcher.Dispatch(model.OrderCreated{
		OrderID:    orderID,
		CustomerID: customerID,
	})
}

func (o *orderService) RemoveOrder(orderID uuid.UUID) error {
	order, err := o.orderRepo.Find(orderID)
	if err != nil {
		if errors.Is(err, model.ErrOrderNotFound) {
			return nil
		}
		return err
	}

	now := time.Now()
	order.DeletedAt = &now
	order.UpdatedAt = now

	if err = o.orderRepo.Store(order); err != nil {
		return err
	}

	return o.eventDispatcher.Dispatch(model.OrderDeleted{
		OrderID: orderID,
	})
}

func (o *orderService) SetStatus(orderID uuid.UUID, status model.OrderStatus) error {
	order, err := o.orderRepo.Find(orderID)
	if err != nil {
		return err
	}

	oldStatus := order.Status

	if !o.isValidStatusTransition(order.Status, status) {
		return ErrInvalidOrderStatus
	}

	order.Status = status
	order.UpdatedAt = time.Now()

	if err = o.orderRepo.Store(order); err != nil {
		return err
	}

	return o.eventDispatcher.Dispatch(model.OrderStatusChanged{
		OrderID: orderID,
		From:    oldStatus,
		To:      status,
	})
}

func (o *orderService) AddItem(orderID, productID uuid.UUID, price float64) (uuid.UUID, error) {
	order, err := o.orderRepo.Find(orderID)
	if err != nil {
		return uuid.Nil, err
	}

	if order.Status != model.Open {
		return uuid.Nil, ErrInvalidOrderStatus
	}

	order.Items = append(order.Items, model.OrderItem{
		OrderID:    orderID,
		ProductID:  productID,
		TotalPrice: price,
	})
	err = o.orderRepo.Store(order)
	if err != nil {
		return uuid.Nil, err
	}

	return productID, o.eventDispatcher.Dispatch(model.OrderItemsChanged{
		OrderID:    orderID,
		AddedItems: []uuid.UUID{productID},
	})
}

func (o *orderService) RemoveItem(orderID, itemID uuid.UUID) error {
	order, err := o.orderRepo.Find(orderID)
	if err != nil {
		return err
	}

	if order.Status != model.Open {
		return ErrInvalidOrderStatus
	}

	found := false
	var newItems []model.OrderItem
	for _, item := range order.Items {
		if item.ProductID != itemID {
			newItems = append(newItems, item)
		} else {
			found = true
		}
	}

	if !found {
		return nil
	}

	order.Items = newItems
	order.UpdatedAt = time.Now()

	if err = o.orderRepo.Store(order); err != nil {
		return err
	}

	return o.eventDispatcher.Dispatch(model.OrderItemsChanged{
		OrderID:      orderID,
		RemovedItems: []uuid.UUID{itemID},
	})
}

func (o *orderService) isValidStatusTransition(from, to model.OrderStatus) bool {
	switch from {
	case model.Open:
		return to == model.Pending || to == model.Cancelled
	case model.Pending:
		return to == model.Paid || to == model.Cancelled
	case model.Paid, model.Cancelled:
		return false
	default:
		return false
	}
}
