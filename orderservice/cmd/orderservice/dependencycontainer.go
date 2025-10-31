package main

import (
	"github.com/jmoiron/sqlx"
)

func newDependencyContainer(
	_ *config,
	connContainer *connectionsContainer,
) (*dependencyContainer, error) {
	return &dependencyContainer{
		DB: connContainer.db,
	}, nil
}

type dependencyContainer struct {
	DB *sqlx.DB
}
