package database

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/migrator"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/pkg/errors"
)

func NewVersion2(client mysql.ClientContext) migrator.Migration {
	return &version2{
		client: client,
	}
}

type version2 struct {
	client mysql.ClientContext
}

func (v version2) Version() int64 {
	return 2
}

func (v version2) Description() string {
	return "Create 'product_reservation' table"
}

func (v version2) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS product_reservation
		(
		    order_id    VARCHAR(64)   NOT NULL,
		    product_id  VARCHAR(64)   NOT NULL,
		    count       INT           NOT NULL,
		    price       DECIMAL(10,2) NOT NULL,
		    created_at  DATETIME      NOT NULL,
		    PRIMARY KEY (order_id, product_id)
		)
		    ENGINE = InnoDB
		    CHARACTER SET = utf8mb4
		    COLLATE utf8mb4_unicode_ci
	`)
	return errors.WithStack(err)
}
