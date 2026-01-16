package database

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/migrator"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/pkg/errors"
)

func NewVersion1(client mysql.ClientContext) migrator.Migration {
	return &version1{
		client: client,
	}
}

type version1 struct {
	client mysql.ClientContext
}

func (v version1) Version() int64 {
	return 1
}

func (v version1) Description() string {
	return "Create 'product' table"
}

func (v version1) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS product
		(
		    product_id  VARCHAR(64)   NOT NULL,
		    name        VARCHAR(255)  NOT NULL,
		    price       DECIMAL(10,2) NOT NULL,
		    stock       INT           NOT NULL,
		    created_at  DATETIME      NOT NULL,
		    updated_at  DATETIME      NOT NULL,
		    deleted_at  DATETIME,
		    PRIMARY KEY (product_id)
		)
		    ENGINE = InnoDB
		    CHARACTER SET = utf8mb4
		    COLLATE utf8mb4_unicode_ci
	`)
	return errors.WithStack(err)
}
