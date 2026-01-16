package service

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/google/uuid"

	commonevent "payment/pkg/common/event"
	"payment/pkg/payment/domain/model"
	domainservice "payment/pkg/payment/domain/service"
)

type PaymentService interface {
	Pay(ctx context.Context, orderID, customerID uuid.UUID, amount float64) error
}

func NewPaymentService(
	uow UnitOfWork,
	luow LockableUnitOfWork,
	eventDispatcher outbox.EventDispatcher[outbox.Event],
) PaymentService {
	return &paymentService{
		uow:             uow,
		luow:            luow,
		eventDispatcher: eventDispatcher,
	}
}

type paymentService struct {
	uow             UnitOfWork
	luow            LockableUnitOfWork
	eventDispatcher outbox.EventDispatcher[outbox.Event]
}

func (s *paymentService) Pay(ctx context.Context, orderID, customerID uuid.UUID, amount float64) error {
	// Лочим по OrderID, чтобы предотвратить дублирующую оплату одного заказа
	err := s.luow.Execute(ctx, []string{paymentLockByOrder(orderID)}, func(provider RepositoryProvider) error {
		// Нам нужны репозитории и кошелька, и платежей
		paymentRepo := provider.PaymentRepository(ctx)
		walletRepo := provider.WalletRepository(ctx)

		domainSvc := s.domainService(ctx, paymentRepo, walletRepo)
		_, err := domainSvc.PerformPayment(orderID, customerID, amount)
		return err
	})

	return err
}

func (s *paymentService) domainService(
	ctx context.Context,
	paymentRepo model.PaymentRepository,
	walletRepo model.WalletRepository,
) domainservice.PaymentService {
	dispatcher := s.domainEventDispatcher(ctx)

	walletDomainSvc := domainservice.NewWalletService(walletRepo, dispatcher)
	return domainservice.NewPaymentService(paymentRepo, walletDomainSvc, dispatcher)
}

func (s *paymentService) domainEventDispatcher(ctx context.Context) commonevent.Dispatcher {
	return &domainEventDispatcher{
		ctx:             ctx,
		eventDispatcher: s.eventDispatcher,
	}
}

const basePaymentLock = "payment_"

func paymentLockByOrder(orderID uuid.UUID) string {
	return basePaymentLock + "order_" + orderID.String()
}
