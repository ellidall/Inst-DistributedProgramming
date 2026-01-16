package integrationevent

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/amqp"
)

var errUnhandledDelivery = errors.New("unhandled delivery")

func NewAMQPTransport(logger logging.Logger) AMQPTransport {
	return &amqpTransport{
		logger: logger,
	}
}

type AMQPTransport interface {
	Handler() amqp.Handler
}

type amqpTransport struct {
	logger logging.Logger
}

func (t *amqpTransport) Handler() amqp.Handler {
	return t.withLog(t.handle)
}

func (t *amqpTransport) handle(_ context.Context, _ amqp.Delivery) error {
	return nil
}

func (t *amqpTransport) withLog(handler amqp.Handler) amqp.Handler {
	return func(ctx context.Context, delivery amqp.Delivery) error {
		l := t.logger.WithFields(logging.Fields{
			"routing_key":    delivery.RoutingKey,
			"correlation_id": delivery.CorrelationID,
			"content_type":   delivery.ContentType,
		})
		if delivery.ContentType != ContentType {
			l.Warning(errors.New("invalid content type"), "skipping")
			return nil
		}
		l = l.WithField("body", json.RawMessage(delivery.Body))

		start := time.Now()
		err := handler(ctx, delivery)
		l.WithField("duration", time.Since(start))

		if err != nil {
			if errors.Is(err, errUnhandledDelivery) {
				l.Info("unhandled delivery, skipping")
				return nil
			}
			l.Error(err, "failed to handle message")
		} else {
			l.Info("successfully handled message")
		}
		return err
	}
}
