package worker

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"order/pkg/order/app/service"
	"order/pkg/order/infrastructure/temporal"
	"order/pkg/order/infrastructure/temporal/activity"
)

func InterruptChannel() <-chan interface{} {
	return worker.InterruptCh()
}

func NewWorker(
	temporalClient client.Client,
	orderService service.OrderService,
) worker.Worker {
	w := worker.New(temporalClient, temporal.OrderTaskQueue, worker.Options{})
	w.RegisterActivity(activity.NewOrderServiceActivities(orderService))
	w.RegisterWorkflow(temporal.ProcessOrderWorkflow)
	return w
}
