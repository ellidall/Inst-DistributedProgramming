package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"product/pkg/product/domain/model"
)

func NewProductRepository(ctx context.Context, client mysql.ClientContext) model.ProductRepository {
	return &productRepository{
		ctx:    ctx,
		client: client,
	}
}

type productRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (r *productRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (r *productRepository) Store(product *model.Product) error {
	_, err := r.client.ExecContext(r.ctx,
		`
		INSERT INTO product (product_id, name, price, stock, created_at, updated_at, deleted_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name=VALUES(name),
			price=VALUES(price),
			stock=VALUES(stock),
			updated_at=VALUES(updated_at),
			deleted_at=VALUES(deleted_at)
		`,
		product.ID,
		product.Name,
		product.Price,
		product.Stock,
		product.CreatedAt,
		product.UpdatedAt,
		toSQLNull(product.DeletedAt),
	)
	return errors.WithStack(err)
}

func (r *productRepository) Find(id uuid.UUID) (*model.Product, error) {
	row := struct {
		ID        uuid.UUID           `db:"product_id"`
		Name      string              `db:"name"`
		Price     float64             `db:"price"`
		Stock     int                 `db:"stock"`
		CreatedAt time.Time           `db:"created_at"`
		UpdatedAt time.Time           `db:"updated_at"`
		DeletedAt sql.Null[time.Time] `db:"deleted_at"`
	}{}

	err := r.client.GetContext(
		r.ctx,
		&row,
		`SELECT product_id, name, price, stock, created_at, updated_at, deleted_at 
		 FROM product 
		 WHERE product_id = ? AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrProductNotFound)
		}
		return nil, errors.WithStack(err)
	}

	return &model.Product{
		ID:        row.ID,
		Name:      row.Name,
		Price:     row.Price,
		Stock:     row.Stock,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		DeletedAt: fromSQLNull(row.DeletedAt),
	}, nil
}

func (r *productRepository) FindMany(ids []uuid.UUID) (map[uuid.UUID]*model.Product, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]*model.Product{}, nil
	}

	// Формируем плейсхолдеры (?,?,?) вручную, чтобы не зависеть от sqlx.In
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(
		`SELECT product_id, name, price, stock, created_at, updated_at, deleted_at 
		 FROM product 
		 WHERE product_id IN (%s) AND deleted_at IS NULL`,
		strings.Join(placeholders, ","),
	)

	var rows []struct {
		ID        uuid.UUID           `db:"product_id"`
		Name      string              `db:"name"`
		Price     float64             `db:"price"`
		Stock     int                 `db:"stock"`
		CreatedAt time.Time           `db:"created_at"`
		UpdatedAt time.Time           `db:"updated_at"`
		DeletedAt sql.Null[time.Time] `db:"deleted_at"`
	}

	err := r.client.SelectContext(r.ctx, &rows, query, args...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := make(map[uuid.UUID]*model.Product, len(rows))
	for _, row := range rows {
		result[row.ID] = &model.Product{
			ID:        row.ID,
			Name:      row.Name,
			Price:     row.Price,
			Stock:     row.Stock,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
			DeletedAt: fromSQLNull(row.DeletedAt),
		}
	}

	return result, nil
}

func (r *productRepository) Remove(id uuid.UUID) error {
	now := time.Now()
	_, err := r.client.ExecContext(r.ctx,
		`UPDATE product SET deleted_at = ? WHERE product_id = ?`,
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
