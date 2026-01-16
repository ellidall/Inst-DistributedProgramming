package activity

import (
	"context"

	"github.com/google/uuid"

	"product/pkg/product/app/data"
	"product/pkg/product/app/service"
)

type ReserveInput struct {
	OrderID uuid.UUID          `json:"order_id"`
	Items   []data.ReserveItem `json:"items"`
}

type ReserveOutput struct {
	TotalPrice float64          `json:"total_price"`
	Items      []data.ItemPrice `json:"items"`
}

func NewProductServiceActivities(productService service.ProductService) *ProductServiceActivities {
	return &ProductServiceActivities{
		productService: productService,
	}
}

type ProductServiceActivities struct {
	productService service.ProductService
}

func (a *ProductServiceActivities) Reserve(ctx context.Context, input ReserveInput) (*ReserveOutput, error) {
	prices, err := a.productService.Reserve(ctx, input.OrderID, input.Items)
	if err != nil {
		return nil, err
	}

	var total float64
	priceMap := make(map[uuid.UUID]float64)
	for _, p := range prices {
		priceMap[p.ProductID] = p.Price
	}

	for _, item := range input.Items {
		if price, ok := priceMap[item.ProductID]; ok {
			total += price * float64(item.Count)
		}
	}

	return &ReserveOutput{
		TotalPrice: total,
		Items:      prices,
	}, nil
}

func (a *ProductServiceActivities) CancelReserve(ctx context.Context, input ReserveInput) error {
	return a.productService.CancelReservation(ctx, input.OrderID)
}
