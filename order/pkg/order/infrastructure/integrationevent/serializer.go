package integrationevent

import (
	"encoding/json"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/pkg/errors"

	"order/pkg/order/domain/model"
)

func NewEventSerializer() outbox.EventSerializer[outbox.Event] {
	return &eventSerializer{}
}

type eventSerializer struct{}

func (s eventSerializer) Serialize(event outbox.Event) (string, error) {
	switch e := event.(type) {
	case model.OrderCreated:
		items := make([]OrderItemDTO, len(e.Items))
		for i, item := range e.Items {
			items[i] = OrderItemDTO{
				ProductID: item.ProductID.String(),
				Count:     item.Count,
			}
		}

		b, err := json.Marshal(OrderCreated{
			OrderID:    e.OrderID.String(),
			CustomerID: e.CustomerID.String(),
			Items:      items,
		})
		return string(b), errors.WithStack(err)

	case model.OrderPaid:
		b, err := json.Marshal(OrderPaid{
			OrderID:     e.OrderID.String(),
			CustomerID:  e.CustomerID.String(),
			TotalAmount: e.TotalAmount,
		})
		return string(b), errors.WithStack(err)

	default:
		return "", errors.Errorf("unknown event %q", event.Type())
	}
}

type OrderItemDTO struct {
	ProductID string `json:"product_id"`
	Count     int    `json:"count"`
}

type OrderCreated struct {
	OrderID    string         `json:"order_id"`
	CustomerID string         `json:"customer_id"`
	Items      []OrderItemDTO `json:"items"`
}

type OrderPaid struct {
	OrderID     string  `json:"order_id"`
	CustomerID  string  `json:"customer_id"`
	TotalAmount float64 `json:"total_amount"`
}
