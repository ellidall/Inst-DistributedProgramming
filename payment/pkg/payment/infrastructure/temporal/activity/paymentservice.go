package activity

import (
	"context"

	"github.com/google/uuid"

	"payment/pkg/payment/app/service"
)

type PayInput struct {
	OrderID    uuid.UUID `json:"order_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	Amount     float64   `json:"amount"`
}

func NewPaymentActivities(paymentService service.PaymentService) *PaymentActivities {
	return &PaymentActivities{
		paymentService: paymentService,
	}
}

type PaymentActivities struct {
	paymentService service.PaymentService
}

func (a *PaymentActivities) Pay(ctx context.Context, input PayInput) error {
	return a.paymentService.Pay(ctx, input.OrderID, input.CustomerID, input.Amount)
}
