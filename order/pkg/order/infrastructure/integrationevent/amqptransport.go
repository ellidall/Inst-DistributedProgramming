package integrationevent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/amqp"

	"order/pkg/order/domain/model"
	"order/pkg/order/infrastructure/temporal"
)

var errUnhandledDelivery = errors.New("unhandled delivery")

func NewAMQPTransport(logger logging.Logger, workflowService temporal.WorkflowService) AMQPTransport {
	return &amqpTransport{
		logger:          logger,
		workflowService: workflowService,
	}
}

type AMQPTransport interface {
	Handler() amqp.Handler
}

type amqpTransport struct {
	logger          logging.Logger
	workflowService temporal.WorkflowService
}

func (t *amqpTransport) Handler() amqp.Handler {
	return t.withLog(t.handle)
}

func (t *amqpTransport) handle(ctx context.Context, delivery amqp.Delivery) error {
	switch delivery.Type {
	case model.OrderCreated{}.Type():
		var e model.OrderCreated
		err := json.Unmarshal(delivery.Body, &e)
		if err != nil {
			return err
		}
		fmt.Println("event = ", e)
		return t.workflowService.RunProcessOrderWorkflow(ctx, delivery.CorrelationID, e)
	default:
		return errUnhandledDelivery
	}
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
