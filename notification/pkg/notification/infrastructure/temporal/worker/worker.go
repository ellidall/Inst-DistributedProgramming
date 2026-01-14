package worker

import (
	"fmt"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"notification/pkg/notification/app/service"
	"notification/pkg/notification/infrastructure/temporal"
	"notification/pkg/notification/infrastructure/temporal/activity"
	"notification/pkg/notification/infrastructure/temporal/workflows"
)

func InterruptChannel() <-chan interface{} {
	return worker.InterruptCh()
}

func NewWorker(
	temporalClient client.Client,
	notificationService service.NotificationService,
	userService service.UserService,
) worker.Worker {
	w := worker.New(temporalClient, temporal.TaskQueue, worker.Options{})
	w.RegisterActivity(activity.NewNotificationActivities(notificationService))
	w.RegisterActivity(activity.NewUserActivities(userService))
	fmt.Println("Registering CreateUserWorkflow...")
	w.RegisterWorkflow(workflows.CreateUserWorkflow)
	fmt.Println("Registered CreateUserWorkflow")
	w.RegisterWorkflow(workflows.UserUpdatedWorkflow)
	return w
}
