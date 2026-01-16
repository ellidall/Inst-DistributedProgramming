package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	commonevent "payment/pkg/common/event"
	"payment/pkg/payment/domain/model"
)

type PaymentService interface {
	PerformPayment(orderID, customerID uuid.UUID, amount float64) (*model.Payment, error)
}

func NewPaymentService(
	paymentRepo model.PaymentRepository,
	walletService WalletService,
	dispatcher commonevent.Dispatcher,
) PaymentService {
	return &paymentService{
		paymentRepo:   paymentRepo,
		walletService: walletService,
		dispatcher:    dispatcher,
	}
}

type paymentService struct {
	paymentRepo   model.PaymentRepository
	walletService WalletService
	dispatcher    commonevent.Dispatcher
}

func (s *paymentService) PerformPayment(orderID, customerID uuid.UUID, amount float64) (*model.Payment, error) {
	paymentID, err := s.paymentRepo.NextID()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// 1. Сначала пытаемся списать деньги через Wallet Service.
	// Это критическая секция. Если тут упадет, платеж даже не создастся (или создастся Failed).

	walletID, err := s.walletService.DeductFunds(customerID, amount)

	var status model.PaymentStatus
	if err != nil {
		// Если денег нет или кошелька нет - платеж Failed
		if errors.Is(err, model.ErrInsufficientFunds) || errors.Is(err, model.ErrWalletNotFound) {
			status = model.Failed
		} else {
			// Техническая ошибка (БД легла) - возвращаем ошибку наверх, пусть Temporal ретраит
			return nil, err
		}
	} else {
		status = model.Succeeded
	}

	// 2. Создаем запись о платеже
	payment := &model.Payment{
		ID:        paymentID,
		WalletID:  walletID,
		OrderID:   orderID,
		Amount:    amount,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.paymentRepo.Store(payment); err != nil {
		return nil, err
	}

	err = s.dispatcher.Dispatch(model.PaymentCreated{
		PaymentID: paymentID,
		WalletID:  walletID,
		OrderID:   orderID,
		Amount:    amount,
	})
	if err != nil {
		return nil, err
	}

	err = s.dispatcher.Dispatch(model.PaymentStatusChanged{
		PaymentID: paymentID,
		From:      model.Pending,
		To:        status,
	})
	if err != nil {
		return nil, err
	}

	if status == model.Failed {
		return payment, errors.New("payment failed: insufficient funds or wallet error")
	}

	return payment, nil
}
