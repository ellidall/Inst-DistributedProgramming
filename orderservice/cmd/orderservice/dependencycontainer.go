package main

import (
	"github.com/jmoiron/sqlx"

	appservice "orderservice/pkg/app/service"
)

func newDependencyContainer(
	_ *config,
	connContainer *connectionsContainer,
) (*dependencyContainer, error) {
	orderService := appservice.NewOrderService()

	return &dependencyContainer{
		DB:           connContainer.db,
		OrderService: orderService,
	}, nil
}

type dependencyContainer struct {
	DB           *sqlx.DB
	OrderService appservice.OrderService
}
