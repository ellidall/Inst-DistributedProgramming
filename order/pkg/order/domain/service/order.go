package service

import (
	"time"

	"github.com/google/uuid"

	commonevent "order/pkg/common/event"
	"order/pkg/order/domain/model"
)

type OrderService interface {
	CreateOrder(orderID, customerID uuid.UUID, items []model.OrderItem) (*model.Order, error)
	UpdateStatus(orderID uuid.UUID, status model.OrderStatus) error
	UpdateItemPrices(orderID uuid.UUID, prices map[uuid.UUID]float64) error
}

func NewOrderService(repo model.OrderRepository, dispatcher commonevent.Dispatcher) OrderService {
	return &orderService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

type orderService struct {
	repo       model.OrderRepository
	dispatcher commonevent.Dispatcher
}

func (s *orderService) CreateOrder(orderID, customerID uuid.UUID, items []model.OrderItem) (*model.Order, error) {
	if len(items) == 0 {
		return nil, model.ErrEmptyOrderItems
	}

	now := time.Now()
	order := &model.Order{
		ID:         orderID,
		CustomerID: customerID,
		Status:     model.Pending,
		Items:      items,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.repo.Store(order); err != nil {
		return nil, err
	}

	eventItems := make([]model.OrderItemDTO, len(items))
	for i, item := range items {
		eventItems[i] = model.OrderItemDTO{
			ProductID: item.ProductID,
			Count:     item.Count,
		}
	}

	err := s.dispatcher.Dispatch(model.OrderCreated{
		OrderID:    orderID,
		CustomerID: customerID,
		Items:      eventItems,
	})
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *orderService) UpdateStatus(orderID uuid.UUID, status model.OrderStatus) error {
	order, err := s.repo.Find(orderID)
	if err != nil {
		return err
	}

	if !isValidTransition(order.Status, status) {
		return model.ErrInvalidOrderStatus
	}

	order.Status = status
	order.UpdatedAt = time.Now()

	if err := s.repo.Store(order); err != nil {
		return err
	}

	if status == model.Paid {
		err := s.dispatcher.Dispatch(model.OrderPaid{
			OrderID:     order.ID,
			CustomerID:  order.CustomerID,
			TotalAmount: order.CalculateTotal(),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *orderService) UpdateItemPrices(orderID uuid.UUID, prices map[uuid.UUID]float64) error {
	order, err := s.repo.Find(orderID)
	if err != nil {
		return err
	}

	if order.Status != model.Pending {
		return model.ErrInvalidOrderStatus
	}

	for i, item := range order.Items {
		if price, ok := prices[item.ProductID]; ok {
			order.Items[i].Price = price
		}
	}
	order.UpdatedAt = time.Now()

	return s.repo.Store(order)
}

func isValidTransition(from, to model.OrderStatus) bool {
	switch from {
	case model.Pending:
		return to == model.Paid || to == model.Cancelled
	default:
		return false
	}
}
