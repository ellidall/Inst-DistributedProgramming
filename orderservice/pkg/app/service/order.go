package service

import (
	"context"
	"orderservice/pkg/app/data"
)

type OrderService interface {
	CreateOrder(ctx context.Context, input data.CreateOrderInput) (int, error)
}

func NewOrderService() OrderService {
	return &orderService{}
}

type orderService struct {
}

func (o orderService) CreateOrder(_ context.Context, _ data.CreateOrderInput) (int, error) {
	return 1, nil
}
