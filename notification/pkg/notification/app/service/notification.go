package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"

	"notification/pkg/notification/app/data"
	"notification/pkg/notification/domain/model"
	"notification/pkg/notification/domain/service"
)

type NotificationService interface {
	CreateNotification(ctx context.Context, payload data.NotificationPayload) (uuid.UUID, error)
	MarkAsExecuted(ctx context.Context, id uuid.UUID, success bool) error
	FindNotification(ctx context.Context, id uuid.UUID) (data.Notification, error)
}

func NewNotificationService(
	uow UnitOfWork,
	luow LockableUnitOfWork,
	eventDispatcher outbox.EventDispatcher[outbox.Event],
) NotificationService {
	return &notificationService{
		uow:             uow,
		luow:            luow,
		eventDispatcher: eventDispatcher,
	}
}

type notificationService struct {
	uow             UnitOfWork
	luow            LockableUnitOfWork
	eventDispatcher outbox.EventDispatcher[outbox.Event]
}

func (s *notificationService) CreateNotification(ctx context.Context, payload data.NotificationPayload) (uuid.UUID, error) {
	fmt.Println("CreateNotification payload = ", payload)
	var id uuid.UUID
	err := s.luow.Execute(ctx, []string{notificationLock(id)}, func(provider RepositoryProvider) error {
		domainPayload := model.NotificationPayload{
			Email:   payload.Email,
			Message: payload.Message,
		}
		domainService := s.notificationDomainService(ctx, provider.NotificationRepository(ctx))
		newID, err := domainService.CreateNotification(domainPayload)
		if err != nil {
			return err
		}
		id = newID
		return nil
	})
	return id, err
}

func (s *notificationService) MarkAsExecuted(ctx context.Context, id uuid.UUID, success bool) error {
	return s.luow.Execute(ctx, []string{notificationLock(id)}, func(provider RepositoryProvider) error {
		return s.notificationDomainService(ctx, provider.NotificationRepository(ctx)).MarkAsExecuted(id, success)
	})
}

func (s *notificationService) FindNotification(ctx context.Context, id uuid.UUID) (data.Notification, error) {
	var dto data.Notification
	err := s.luow.Execute(ctx, []string{notificationLock(id)}, func(provider RepositoryProvider) error {
		domainNotif, err := provider.NotificationRepository(ctx).Find(id)
		if err != nil {
			return err
		}
		dto = data.Notification{
			ID:         domainNotif.ID,
			Payload:    data.NotificationPayload(domainNotif.Payload),
			ExecutedAt: domainNotif.ExecutedAt,
			CreatedAt:  domainNotif.CreatedAt,
			UpdatedAt:  domainNotif.UpdatedAt,
			Status:     data.NotificationStatusFromDomain(domainNotif.Status),
		}
		return nil
	})
	return dto, err
}

func (s *notificationService) notificationDomainService(
	_ context.Context,
	repo model.NotificationRepository,
) service.NotificationService {
	return service.NewNotificationService(repo)
}

const baseNotificationLock = "notification_"

func notificationLock(id uuid.UUID) string {
	return baseNotificationLock + id.String()
}
