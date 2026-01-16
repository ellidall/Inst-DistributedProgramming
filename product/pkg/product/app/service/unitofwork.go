package service

import (
	"context"

	"product/pkg/product/domain/model"
)

type RepositoryProvider interface {
	ProductRepository(ctx context.Context) model.ProductRepository
	ReservationRepository(ctx context.Context) model.ReservationRepository
}

type LockableUnitOfWork interface {
	Execute(ctx context.Context, lockNames []string, f func(provider RepositoryProvider) error) error
}
type UnitOfWork interface {
	Execute(ctx context.Context, f func(provider RepositoryProvider) error) error
}
