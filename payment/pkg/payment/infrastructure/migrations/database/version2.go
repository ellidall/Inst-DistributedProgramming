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
	return "Create 'payment' table"
}

func (v version2) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS payment
		(
		    payment_id VARCHAR(64)  NOT NULL,
		    wallet_id  VARCHAR(64)  NOT NULL,
		    order_id   VARCHAR(64)  NOT NULL,
		    amount     DECIMAL(15,2) NOT NULL,
		    status     INT          NOT NULL,
		    created_at DATETIME     NOT NULL,
		    updated_at DATETIME     NOT NULL,
		    deleted_at DATETIME,
		    PRIMARY KEY (payment_id)
		)
		    ENGINE = InnoDB
		    CHARACTER SET = utf8mb4
		    COLLATE utf8mb4_unicode_ci
	`)
	return errors.WithStack(err)
}
