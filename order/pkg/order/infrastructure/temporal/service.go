package temporal

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"

	"order/pkg/order/domain/model"
)

const (
	OrderTaskQueue   = "order_task_queue"
	ProductTaskQueue = "product_task_queue"
	PaymentTaskQueue = "payment_task_queue"
)

type WorkflowService interface {
	RunProcessOrderWorkflow(ctx context.Context, id string, event model.OrderCreated) error
}

func NewWorkflowService(temporalClient client.Client) WorkflowService {
	return &workflowService{
		temporalClient: temporalClient,
	}
}

type workflowService struct {
	temporalClient client.Client
}

func (s *workflowService) RunProcessOrderWorkflow(ctx context.Context, id string, event model.OrderCreated) error {
	fmt.Println("RunProcessOrderWorkflow event = ", event)
	_, err := s.temporalClient.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        id,
			TaskQueue: OrderTaskQueue,
		},
		ProcessOrderWorkflow, event,
	)
	return err
}
