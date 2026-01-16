package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	commonevent "payment/pkg/common/event"
	"payment/pkg/payment/domain/model"
)

const defaultBalance = 100000.0

var (
	ErrInvalidWalletBalance = errors.New("invalid wallet balance")
)

type WalletService interface {
	CreateWallet(userID uuid.UUID) (uuid.UUID, error)
	RemoveWallet(walletID uuid.UUID) error
	UpdateWalletBalance(walletID uuid.UUID, newBalance float64) error
	DeductFunds(userID uuid.UUID, amount float64) (uuid.UUID, error)
}

func NewWalletService(repo model.WalletRepository, dispatcher commonevent.Dispatcher) WalletService {
	return &walletService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

type walletService struct {
	repo       model.WalletRepository
	dispatcher commonevent.Dispatcher
}

func (w walletService) CreateWallet(userID uuid.UUID) (uuid.UUID, error) {
	walletID, err := w.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	initialBalance := defaultBalance
	currentTime := time.Now()
	err = w.repo.Store(&model.Wallet{
		ID:        walletID,
		UserID:    userID,
		Balance:   initialBalance,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return walletID, w.dispatcher.Dispatch(model.WalletCreated{
		WalletID: walletID,
		UserID:   userID,
		Balance:  initialBalance,
	})
}

func (w walletService) RemoveWallet(walletID uuid.UUID) error {
	wallet, err := w.repo.Find(walletID)
	if err != nil {
		if errors.Is(err, model.ErrWalletNotFound) {
			return nil
		}
		return err
	}

	now := time.Now()
	wallet.DeletedAt = &now
	wallet.UpdatedAt = now

	if err = w.repo.Store(wallet); err != nil {
		return err
	}

	return w.dispatcher.Dispatch(model.WalletRemoved{
		WalletID: walletID,
	})
}

func (w walletService) UpdateWalletBalance(walletID uuid.UUID, newBalance float64) error {
	wallet, err := w.repo.Find(walletID)
	if err != nil {
		return err
	}

	oldBalance := wallet.Balance

	if newBalance < 0 {
		return ErrInvalidWalletBalance
	}

	wallet.Balance = newBalance
	wallet.UpdatedAt = time.Now()

	if err = w.repo.Store(wallet); err != nil {
		return err
	}

	return w.dispatcher.Dispatch(model.WalletBalanceChanged{
		WalletID:   walletID,
		OldBalance: oldBalance,
		NewBalance: newBalance,
	})
}

func (w walletService) DeductFunds(userID uuid.UUID, amount float64) (uuid.UUID, error) {
	// 1. Ищем кошелек пользователя
	wallet, err := w.repo.FindByUserID(userID)
	if err != nil {
		return uuid.Nil, err
	}

	// 2. Проверяем баланс
	if wallet.Balance < amount {
		return wallet.ID, model.ErrInsufficientFunds
	}

	oldBalance := wallet.Balance
	newBalance := wallet.Balance - amount

	// 3. Обновляем баланс
	wallet.Balance = newBalance
	wallet.UpdatedAt = time.Now()

	if err := w.repo.Store(wallet); err != nil {
		return wallet.ID, err
	}

	// 4. Отправляем событие изменения баланса
	err = w.dispatcher.Dispatch(model.WalletBalanceChanged{
		WalletID:   wallet.ID,
		OldBalance: oldBalance,
		NewBalance: newBalance,
	})

	return wallet.ID, err
}
