package workflows

import (
	"time"

	"go.temporal.io/sdk/workflow"

	appdata "notification/pkg/notification/app/data"
	"notification/pkg/notification/domain/model"
)

func UserUpdatedWorkflow(ctx workflow.Context, event model.UserUpdated) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("UserUpdatedWorkflow started", "user_id", event.UserID)

	// Проверяем, есть ли вообще изменения контактных данных
	hasContactChange := false

	if event.UpdatedFields != nil {
		if event.UpdatedFields.Email != nil || event.UpdatedFields.Telegram != nil {
			hasContactChange = true
		}
	}
	if event.RemovedFields != nil {
		if (event.RemovedFields.Email != nil && *event.RemovedFields.Email) ||
			(event.RemovedFields.Telegram != nil && *event.RemovedFields.Telegram) {
			hasContactChange = true
		}
	}

	if !hasContactChange {
		logger.Info("No relevant contact fields changed, skipping")
		return nil
	}

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})

	// Получаем текущего пользователя для определения нового статуса
	var user appdata.User
	err := workflow.ExecuteActivity(ctx, userActivities.FindUser, event.UserID).Get(ctx, &user)
	if err != nil {
		return err
	}

	// Формируем update на основе события
	update := appdata.UserUpdate{}

	// Обработка UpdatedFields
	if event.UpdatedFields != nil {
		if event.UpdatedFields.Status != nil {
			status := int(*event.UpdatedFields.Status)
			update.Status = &status
		}
		update.Email = event.UpdatedFields.Email
		update.Telegram = event.UpdatedFields.Telegram
	}

	// Обработка RemovedFields: устанавливаем в nil
	if event.RemovedFields != nil {
		if event.RemovedFields.Email != nil && *event.RemovedFields.Email {
			emailNil := ""
			update.Email = &emailNil // или nil? зависит от логики
			// Но в твоём случае — лучше передавать nil, чтобы удалить
			update.Email = nil
		}
		if event.RemovedFields.Telegram != nil && *event.RemovedFields.Telegram {
			update.Telegram = nil
		}
	}

	// Выполняем обновление
	err = workflow.ExecuteActivity(ctx, userActivities.UpdateUser, event.UserID, update).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to update user", "error", err)
		return err
	}

	// Отправляем уведомление, только если email был установлен (не удалён)
	if update.Email != nil && *update.Email != "" {
		payload := appdata.NotificationPayload{
			Email:   *update.Email,
			Message: "Your profile has been updated",
		}
		logger.Info("Sending notification", "email", *update.Email)
		err = workflow.ExecuteActivity(ctx, notificationActivities.CreateNotification, payload).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to send notification", "error", err)
			return err // или логировать и продолжать
		}
	}

	return nil
}
