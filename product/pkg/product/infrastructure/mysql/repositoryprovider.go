package mysql

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"

	"product/pkg/product/app/service"
	"product/pkg/product/domain/model"
	"product/pkg/product/infrastructure/mysql/repository"
)

func NewRepositoryProvider(client mysql.ClientContext) service.RepositoryProvider {
	return &repositoryProvider{client: client}
}

type repositoryProvider struct {
	client mysql.ClientContext
}

func (r *repositoryProvider) ProductRepository(ctx context.Context) model.ProductRepository {
	return repository.NewProductRepository(ctx, r.client)
}

func (r *repositoryProvider) ReservationRepository(ctx context.Context) model.ReservationRepository {
	return repository.NewReservationRepository(ctx, r.client)
}
