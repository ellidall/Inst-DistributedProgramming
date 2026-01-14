package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"payment/pkg/payment/domain/model"
)

func NewWalletRepository(ctx context.Context, client mysql.ClientContext) model.WalletRepository {
	return &walletRepository{
		ctx:    ctx,
		client: client,
	}
}

type walletRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (w *walletRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (w *walletRepository) Store(wallet *model.Wallet) error {
	fmt.Println("Store wallet = ", wallet)
	_, err := w.client.ExecContext(w.ctx,
		`
	INSERT INTO wallet (wallet_id, user_id, balance, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		user_id=VALUES(user_id),
		balance=VALUES(balance),
		updated_at=VALUES(updated_at),
		deleted_at=VALUES(deleted_at)
	`,
		wallet.ID,
		wallet.UserID,
		wallet.Balance,
		wallet.CreatedAt,
		wallet.UpdatedAt,
		toSQLNull(wallet.DeletedAt),
	)
	return errors.WithStack(err)
}

func (w *walletRepository) Find(id uuid.UUID) (*model.Wallet, error) {
	walletRow := struct {
		ID        uuid.UUID           `db:"wallet_id"`
		UserID    uuid.UUID           `db:"user_id"`
		Balance   float64             `db:"balance"`
		CreatedAt time.Time           `db:"created_at"`
		UpdatedAt time.Time           `db:"updated_at"`
		DeletedAt sql.Null[time.Time] `db:"deleted_at"`
	}{}

	err := w.client.GetContext(
		w.ctx,
		&walletRow,
		`SELECT wallet_id, user_id, balance, created_at, updated_at, deleted_at FROM wallet WHERE wallet_id = ?`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrWalletNotFound)
		}
		return nil, errors.WithStack(err)
	}

	return &model.Wallet{
		ID:        walletRow.ID,
		UserID:    walletRow.UserID,
		Balance:   walletRow.Balance,
		CreatedAt: walletRow.CreatedAt,
		UpdatedAt: walletRow.UpdatedAt,
		DeletedAt: fromSQLNull(walletRow.DeletedAt),
	}, nil
}

func (w *walletRepository) Remove(id uuid.UUID) error {
	now := time.Now()
	_, err := w.client.ExecContext(w.ctx,
		`UPDATE wallet SET deleted_at = ? WHERE wallet_id = ?`,
		now,
		id,
	)
	return errors.WithStack(err)
}

func fromSQLNull[T any](v sql.Null[T]) *T {
	if v.Valid {
		return &v.V
	}
	return nil
}

func toSQLNull[T any](v *T) sql.Null[T] {
	if v == nil {
		return sql.Null[T]{}
	}
	return sql.Null[T]{
		V:     *v,
		Valid: true,
	}
}
