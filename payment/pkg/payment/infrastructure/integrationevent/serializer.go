package integrationevent

import (
	"encoding/json"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/pkg/errors"

	"payment/pkg/payment/domain/model"
)

func NewEventSerializer() outbox.EventSerializer[outbox.Event] {
	return &eventSerializer{}
}

type eventSerializer struct{}

func (s eventSerializer) Serialize(event outbox.Event) (string, error) {
	switch e := event.(type) {
	case model.WalletCreated:
		b, err := json.Marshal(WalletCreated{
			WalletID: e.WalletID.String(),
			UserID:   e.UserID.String(),
			Balance:  e.Balance,
		})
		return string(b), errors.WithStack(err)

	case model.WalletBalanceChanged:
		b, err := json.Marshal(WalletBalanceChanged{
			WalletID:   e.WalletID.String(),
			OldBalance: e.OldBalance,
			NewBalance: e.NewBalance,
		})
		return string(b), errors.WithStack(err)

	case model.WalletRemoved:
		b, err := json.Marshal(WalletRemoved{
			WalletID: e.WalletID.String(),
		})
		return string(b), errors.WithStack(err)

	case model.PaymentCreated:
		b, err := json.Marshal(PaymentCreated{
			PaymentID: e.PaymentID.String(),
			WalletID:  e.WalletID.String(),
			OrderID:   e.OrderID.String(),
			Amount:    e.Amount,
		})
		return string(b), errors.WithStack(err)

	case model.PaymentStatusChanged:
		b, err := json.Marshal(PaymentStatusChanged{
			PaymentID: e.PaymentID.String(),
			From:      int(e.From),
			To:        int(e.To),
		})
		return string(b), errors.WithStack(err)

	case model.PaymentRemoved:
		b, err := json.Marshal(PaymentRemoved{
			PaymentID: e.PaymentID.String(),
		})
		return string(b), errors.WithStack(err)

	default:
		return "", errors.Errorf("unknown event %q", event.Type())
	}
}

type WalletCreated struct {
	WalletID string  `json:"wallet_id"`
	UserID   string  `json:"user_id"`
	Balance  float64 `json:"balance"`
}

type WalletBalanceChanged struct {
	WalletID   string  `json:"wallet_id"`
	OldBalance float64 `json:"old_balance"`
	NewBalance float64 `json:"new_balance"`
}

type WalletRemoved struct {
	WalletID string `json:"wallet_id"`
}

type PaymentCreated struct {
	PaymentID string  `json:"payment_id"`
	WalletID  string  `json:"wallet_id"`
	OrderID   string  `json:"order_id"`
	Amount    float64 `json:"amount"`
}

type PaymentStatusChanged struct {
	PaymentID string `json:"payment_id"`
	From      int    `json:"from"`
	To        int    `json:"to"`
}

type PaymentRemoved struct {
	PaymentID string `json:"payment_id"`
}
