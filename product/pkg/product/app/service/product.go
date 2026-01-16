package service

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/google/uuid"

	commonevent "product/pkg/common/event"
	appdata "product/pkg/product/app/data"
	"product/pkg/product/domain/model"
	domainservice "product/pkg/product/domain/service"
)

type ProductService interface {
	CreateProduct(ctx context.Context, name string, price float64, stock int) (uuid.UUID, error)
	Reserve(ctx context.Context, orderID uuid.UUID, items []appdata.ReserveItem) ([]appdata.ItemPrice, error)
	CancelReservation(ctx context.Context, orderID uuid.UUID) error
	GetProduct(ctx context.Context, id uuid.UUID) (appdata.Product, error)
}

func NewProductService(
	uow UnitOfWork,
	luow LockableUnitOfWork,
	eventDispatcher outbox.EventDispatcher[outbox.Event],
) ProductService {
	return &productService{
		uow:             uow,
		luow:            luow,
		eventDispatcher: eventDispatcher,
	}
}

type productService struct {
	uow             UnitOfWork
	luow            LockableUnitOfWork
	eventDispatcher outbox.EventDispatcher[outbox.Event]
}

func (s *productService) CreateProduct(ctx context.Context, name string, price float64, stock int) (uuid.UUID, error) {
	var productID uuid.UUID
	err := s.uow.Execute(ctx, func(provider RepositoryProvider) error {
		domainSvc := s.domainService(ctx, provider.ProductRepository(ctx), provider.ReservationRepository(ctx))
		product, err := domainSvc.CreateProduct(name, price, stock)
		if err != nil {
			return err
		}
		productID = product.ID
		return nil
	})
	return productID, err
}

func (s *productService) Reserve(ctx context.Context, orderID uuid.UUID, items []appdata.ReserveItem) ([]appdata.ItemPrice, error) {
	var prices []appdata.ItemPrice

	lockKeys := make([]string, len(items)+1)
	lockKeys[0] = reservationLock(orderID)

	domainItems := make([]domainservice.ItemRequest, len(items))
	for i, item := range items {
		domainItems[i] = domainservice.ItemRequest{
			ProductID: item.ProductID,
			Count:     item.Count,
		}
		lockKeys[i+1] = productLock(item.ProductID)
	}

	err := s.luow.Execute(ctx, lockKeys, func(provider RepositoryProvider) error {
		domainSvc := s.domainService(ctx, provider.ProductRepository(ctx), provider.ReservationRepository(ctx))

		domainPrices, err := domainSvc.ReserveStock(orderID, domainItems)
		if err != nil {
			return err
		}

		prices = make([]appdata.ItemPrice, len(domainPrices))
		for i, p := range domainPrices {
			prices[i] = appdata.ItemPrice{
				ProductID: p.ProductID,
				Price:     p.Price,
			}
		}
		return nil
	})

	return prices, err
}

func (s *productService) CancelReservation(ctx context.Context, orderID uuid.UUID) error {
	return s.luow.Execute(ctx, []string{reservationLock(orderID)}, func(provider RepositoryProvider) error {
		// В идеале нужно лочить и продукты, но для отмены достаточно лока резервации
		// или оптимистичной блокировки внутри домена
		domainSvc := s.domainService(ctx, provider.ProductRepository(ctx), provider.ReservationRepository(ctx))
		return domainSvc.CancelReservation(orderID)
	})
}

func (s *productService) GetProduct(ctx context.Context, id uuid.UUID) (appdata.Product, error) {
	var res appdata.Product
	err := s.uow.Execute(ctx, func(provider RepositoryProvider) error {
		product, err := provider.ProductRepository(ctx).Find(id)
		if err != nil {
			return err
		}
		res = appdata.Product{
			ID:    product.ID,
			Name:  product.Name,
			Price: product.Price,
			Stock: product.Stock,
		}
		return nil
	})
	return res, err
}

func (s *productService) domainService(ctx context.Context, repository model.ProductRepository, reservationRepository model.ReservationRepository) domainservice.ProductService {
	return domainservice.NewProductService(repository, reservationRepository, s.domainEventDispatcher(ctx))
}

func (s *productService) domainEventDispatcher(ctx context.Context) commonevent.Dispatcher {
	return &domainEventDispatcher{
		ctx:             ctx,
		eventDispatcher: s.eventDispatcher,
	}
}

const baseProductLock = "product_"
const baseReservationLock = "reservation_"

func productLock(id uuid.UUID) string {
	return baseProductLock + id.String()
}

func reservationLock(id uuid.UUID) string {
	return baseReservationLock + id.String()
}
