package activity

import (
	"context"

	"github.com/google/uuid"

	"order/pkg/order/app/service"
)

func NewOrderServiceActivities(orderService service.OrderService) *OrderServiceActivities {
	return &OrderServiceActivities{
		orderService: orderService,
	}
}

type OrderServiceActivities struct {
	orderService service.OrderService
}

func (a *OrderServiceActivities) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status int) error {
	return a.orderService.UpdateOrderStatus(ctx, orderID, status)
}

func (a *OrderServiceActivities) UpdatePrices(ctx context.Context, orderID uuid.UUID, prices []service.ItemPriceUpdate) error {
	return a.orderService.UpdatePrices(ctx, orderID, prices)
}
