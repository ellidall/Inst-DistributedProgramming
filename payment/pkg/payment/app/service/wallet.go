package service

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/google/uuid"

	commonevent "payment/pkg/common/event"
	"payment/pkg/payment/app/data"
	"payment/pkg/payment/domain/model"
	"payment/pkg/payment/domain/service"
)

type WalletService interface {
	CreateWallet(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	RemoveWallet(ctx context.Context, walletID uuid.UUID) error
	UpdateWalletBalance(ctx context.Context, walletID uuid.UUID, newBalance float64) error
	FindWallet(ctx context.Context, walletID uuid.UUID) (data.Wallet, error)
}

func NewWalletService(
	uow UnitOfWork,
	luow LockableUnitOfWork,
	eventDispatcher outbox.EventDispatcher[outbox.Event],
) WalletService {
	return &walletService{
		uow:             uow,
		luow:            luow,
		eventDispatcher: eventDispatcher,
	}
}

type walletService struct {
	uow             UnitOfWork
	luow            LockableUnitOfWork
	eventDispatcher outbox.EventDispatcher[outbox.Event]
}

func (s *walletService) CreateWallet(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	var walletID uuid.UUID

	err := s.luow.Execute(ctx, []string{walletLockByUser(userID)}, func(provider RepositoryProvider) error {
		domainService := s.walletDomainService(ctx, provider.WalletRepository(ctx))
		id, err := domainService.CreateWallet(userID)
		if err != nil {
			return err
		}
		walletID = id
		return nil
	})

	return walletID, err
}

func (s *walletService) RemoveWallet(ctx context.Context, walletID uuid.UUID) error {
	return s.luow.Execute(ctx, []string{walletLock(walletID)}, func(provider RepositoryProvider) error {
		return s.walletDomainService(ctx, provider.WalletRepository(ctx)).RemoveWallet(walletID)
	})
}

func (s *walletService) UpdateWalletBalance(ctx context.Context, walletID uuid.UUID, newBalance float64) error {
	return s.luow.Execute(ctx, []string{walletLock(walletID)}, func(provider RepositoryProvider) error {
		return s.walletDomainService(ctx, provider.WalletRepository(ctx)).UpdateWalletBalance(walletID, newBalance)
	})
}

func (s *walletService) FindWallet(ctx context.Context, walletID uuid.UUID) (data.Wallet, error) {
	var wallet data.Wallet
	err := s.luow.Execute(ctx, []string{walletLock(walletID)}, func(provider RepositoryProvider) error {
		domainWallet, err := provider.WalletRepository(ctx).Find(walletID)
		if err != nil {
			return err
		}
		wallet = data.Wallet{
			ID:      domainWallet.ID,
			UserID:  domainWallet.UserID,
			Balance: domainWallet.Balance,
		}
		return nil
	})
	return wallet, err
}

func (s *walletService) walletDomainService(ctx context.Context, repository model.WalletRepository) service.WalletService {
	return service.NewWalletService(repository, s.domainEventDispatcher(ctx))
}

func (s *walletService) domainEventDispatcher(ctx context.Context) commonevent.Dispatcher {
	return &domainEventDispatcher{
		ctx:             ctx,
		eventDispatcher: s.eventDispatcher,
	}
}

const baseWalletLock = "wallet_"

func walletLock(id uuid.UUID) string {
	return baseWalletLock + id.String()
}

func walletLockByUser(userID uuid.UUID) string {
	return baseWalletLock + "user_" + userID.String()
}
