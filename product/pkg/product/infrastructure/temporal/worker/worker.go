package worker

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"product/pkg/product/app/service"
	"product/pkg/product/infrastructure/temporal/activity"
)

const TaskQueue = "product_task_queue"

func InterruptChannel() <-chan interface{} {
	return worker.InterruptCh()
}

func NewWorker(
	temporalClient client.Client,
	productService service.ProductService,
) worker.Worker {
	w := worker.New(temporalClient, TaskQueue, worker.Options{})
	w.RegisterActivity(activity.NewProductServiceActivities(productService))
	return w
}
