package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	appdata "notification/pkg/notification/app/data"
	"notification/pkg/notification/domain/model"
	"notification/pkg/notification/infrastructure/temporal/activity"
)

var userActivities *activity.UserActivities
var notificationActivities *activity.NotificationActivities

func CreateUserWorkflow(ctx workflow.Context, event model.UserCreated) error {
	fmt.Println("CreateUserWorkflow event = ", event)
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})

	user := appdata.User{
		UserID:   event.UserID,
		Status:   int(event.Status),
		Login:    event.Login,
		Email:    event.Email,
		Telegram: event.Telegram,
	}

	err := workflow.ExecuteActivity(ctx, userActivities.CreateUser, user).Get(ctx, nil)
	if err != nil {
		return err
	}

	if event.Email == nil {
		return nil
	}

	payload := appdata.NotificationPayload{
		Email:   *event.Email,
		Message: "User Created",
	}
	fmt.Println("payload = ", payload)
	return workflow.ExecuteActivity(ctx, notificationActivities.CreateNotification, payload).Get(ctx, nil)
}
