package service

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/google/uuid"

	commonevent "order/pkg/common/event"
	appdata "order/pkg/order/app/data"
	"order/pkg/order/domain/model"
	domainservice "order/pkg/order/domain/service"
)

type ItemPriceUpdate struct {
	ProductID uuid.UUID
	Price     float64
}

type OrderService interface {
	CreateOrder(ctx context.Context, customerID uuid.UUID, items []appdata.OrderItem) (uuid.UUID, error)
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status int) error
	UpdatePrices(ctx context.Context, orderID uuid.UUID, prices []ItemPriceUpdate) error
	FindOrder(ctx context.Context, orderID uuid.UUID) (appdata.Order, error)
}

func NewOrderService(
	uow UnitOfWork,
	luow LockableUnitOfWork,
	eventDispatcher outbox.EventDispatcher[outbox.Event],
) OrderService {
	return &orderService{
		uow:             uow,
		luow:            luow,
		eventDispatcher: eventDispatcher,
	}
}

type orderService struct {
	uow             UnitOfWork
	luow            LockableUnitOfWork
	eventDispatcher outbox.EventDispatcher[outbox.Event]
}

func (s *orderService) CreateOrder(ctx context.Context, customerID uuid.UUID, items []appdata.OrderItem) (uuid.UUID, error) {
	orderID := uuid.New()

	err := s.luow.Execute(ctx, []string{orderLock(orderID)}, func(provider RepositoryProvider) error {
		domainItems := make([]model.OrderItem, len(items))
		for i, item := range items {
			domainItems[i] = model.OrderItem{
				OrderID:   orderID,
				ProductID: item.ProductID,
				Count:     item.Count,
				Price:     0,
			}
		}

		domainSvc := s.domainService(ctx, provider.OrderRepository(ctx))
		_, err := domainSvc.CreateOrder(orderID, customerID, domainItems)
		return err
	})

	if err != nil {
		return uuid.Nil, err
	}
	return orderID, nil
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status int) error {
	newStatus := model.OrderStatus(status)

	return s.luow.Execute(ctx, []string{orderLock(orderID)}, func(provider RepositoryProvider) error {
		domainSvc := s.domainService(ctx, provider.OrderRepository(ctx))
		return domainSvc.UpdateStatus(orderID, newStatus)
	})
}

func (s *orderService) UpdatePrices(ctx context.Context, orderID uuid.UUID, prices []ItemPriceUpdate) error {
	return s.luow.Execute(ctx, []string{orderLock(orderID)}, func(provider RepositoryProvider) error {
		priceMap := make(map[uuid.UUID]float64)
		for _, p := range prices {
			priceMap[p.ProductID] = p.Price
		}

		domainSvc := s.domainService(ctx, provider.OrderRepository(ctx))
		return domainSvc.UpdateItemPrices(orderID, priceMap)
	})
}

func (s *orderService) FindOrder(ctx context.Context, orderID uuid.UUID) (appdata.Order, error) {
	var res appdata.Order
	err := s.uow.Execute(ctx, func(provider RepositoryProvider) error {
		order, err := provider.OrderRepository(ctx).Find(orderID)
		if err != nil {
			return err
		}

		outItems := make([]appdata.OrderItem, len(order.Items))
		for i, item := range order.Items {
			outItems[i] = appdata.OrderItem{
				OrderID:    item.OrderID,
				ProductID:  item.ProductID,
				Count:      item.Count,
				TotalPrice: item.Price,
			}
		}

		res = appdata.Order{
			ID:         order.ID,
			CustomerID: order.CustomerID,
			Status:     appdata.OrderStatus(order.Status),
			Items:      outItems,
			CreatedAt:  order.CreatedAt,
			UpdatedAt:  order.UpdatedAt,
			DeletedAt:  order.DeletedAt,
		}
		return nil
	})
	return res, err
}

func (s *orderService) domainService(ctx context.Context, repository model.OrderRepository) domainservice.OrderService {
	return domainservice.NewOrderService(repository, s.domainEventDispatcher(ctx))
}

func (s *orderService) domainEventDispatcher(ctx context.Context) commonevent.Dispatcher {
	return &domainEventDispatcher{
		ctx:             ctx,
		eventDispatcher: s.eventDispatcher,
	}
}

const baseOrderLock = "order_"

func orderLock(id uuid.UUID) string {
	return baseOrderLock + id.String()
}
