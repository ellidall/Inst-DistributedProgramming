package integrationevent

import (
	"encoding/json"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/pkg/errors"

	"product/pkg/product/domain/model"
)

func NewEventSerializer() outbox.EventSerializer[outbox.Event] {
	return &eventSerializer{}
}

type eventSerializer struct{}

func (s eventSerializer) Serialize(event outbox.Event) (string, error) {
	switch e := event.(type) {
	case model.ProductCreated:
		b, err := json.Marshal(ProductCreated{
			ProductID: e.ProductID.String(),
			Name:      e.Name,
			Price:     e.Price,
			Stock:     e.Stock,
		})
		return string(b), errors.WithStack(err)

	case model.ProductUpdated:
		b, err := json.Marshal(ProductUpdated{
			ProductID: e.ProductID.String(),
			Name:      e.Name,
			Price:     e.Price,
			Stock:     e.Stock,
		})
		return string(b), errors.WithStack(err)

	case model.ProductRemoved:
		b, err := json.Marshal(ProductRemoved{
			ProductID: e.ProductID.String(),
		})
		return string(b), errors.WithStack(err)

	default:
		return "", errors.Errorf("unknown event %q", event.Type())
	}
}

type ProductCreated struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Stock     int     `json:"stock"`
}

type ProductUpdated struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Stock     int     `json:"stock"`
}

type ProductRemoved struct {
	ProductID string `json:"product_id"`
}
