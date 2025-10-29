package transport

import (
	"context"

	api "orderservice/api/server/orderserviceinternal"
	"orderservice/pkg/app/data"
	appservice "orderservice/pkg/app/service"
)

func NewInternalAPI(
	orderService appservice.OrderService,
) api.OrderServiceInternalServiceServer {
	return &internalAPI{
		orderService: orderService,
	}
}

type internalAPI struct {
	orderService appservice.OrderService
}

func (i *internalAPI) CreateOrder(ctx context.Context, _ *api.CreateOrderRequest) (*api.CreateOrderResponse, error) {
	id, err := i.orderService.CreateOrder(ctx, data.CreateOrderInput{
		ProductID: 1,
		Count:     1,
	})
	if err != nil {
		return nil, err
	}

	return &api.CreateOrderResponse{
		OrderId: int64(id),
	}, nil
}
