package repository

import (
	"context"
	"database/sql"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"order/pkg/order/domain/model"
)

func NewOrderRepository(ctx context.Context, client mysql.ClientContext) model.OrderRepository {
	return &orderRepository{
		ctx:    ctx,
		client: client,
	}
}

type orderRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (o *orderRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (o *orderRepository) Store(order *model.Order) error {
	_, err := o.client.ExecContext(o.ctx,
		`
	INSERT INTO orders (order_id, customer_id, status, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		customer_id=VALUES(customer_id),
		status=VALUES(status),
		updated_at=VALUES(updated_at),
		deleted_at=VALUES(deleted_at)
	`,
		order.ID,
		order.CustomerID,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
		toSQLNull(order.DeletedAt),
	)
	if err != nil {
		return errors.WithStack(err)
	}

	// Удаляем старые позиции и вставляем новые
	if len(order.Items) > 0 {
		_, err = o.client.ExecContext(o.ctx, `DELETE FROM order_items WHERE order_id = ?`, order.ID)
		if err != nil {
			return errors.WithStack(err)
		}

		for _, item := range order.Items {
			_, err = o.client.ExecContext(o.ctx,
				`
				INSERT INTO order_items (order_id, product_id, count, total_price) VALUES (?, ?, ?, ?)
				ON DUPLICATE KEY UPDATE
					product_id=VALUES(product_id),
					count=VALUES(count),
					total_price=VALUES(total_price)
				`,
				item.OrderID,
				item.ProductID,
				item.Count,
				item.Price,
			)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}

func (o *orderRepository) Find(id uuid.UUID) (*model.Order, error) {
	orderRow := struct {
		ID         uuid.UUID           `db:"order_id"`
		CustomerID uuid.UUID           `db:"customer_id"`
		Status     int                 `db:"status"`
		CreatedAt  time.Time           `db:"created_at"`
		UpdatedAt  time.Time           `db:"updated_at"`
		DeletedAt  sql.Null[time.Time] `db:"deleted_at"`
	}{}

	err := o.client.GetContext(
		o.ctx,
		&orderRow,
		`SELECT order_id, customer_id, status, created_at, updated_at, deleted_at FROM orders WHERE order_id = ?`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrOrderNotFound)
		}
		return nil, errors.WithStack(err)
	}

	items, err := o.loadOrderItems(orderRow.ID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &model.Order{
		ID:         orderRow.ID,
		CustomerID: orderRow.CustomerID,
		Status:     model.OrderStatus(orderRow.Status),
		Items:      items,
		CreatedAt:  orderRow.CreatedAt,
		UpdatedAt:  orderRow.UpdatedAt,
		DeletedAt:  fromSQLNull(orderRow.DeletedAt),
	}, nil
}

func (o *orderRepository) Remove(id uuid.UUID) error {
	now := time.Now()
	_, err := o.client.ExecContext(o.ctx,
		`UPDATE orders SET deleted_at = ? WHERE order_id = ?`,
		now,
		id,
	)
	return errors.WithStack(err)
}

func (o *orderRepository) loadOrderItems(orderID uuid.UUID) ([]model.OrderItem, error) {
	var itemRows []struct {
		OrderID    uuid.UUID `db:"order_id"`
		ProductID  uuid.UUID `db:"product_id"`
		Count      int       `db:"count"`
		TotalPrice float64   `db:"total_price"`
	}

	err := o.client.SelectContext(
		o.ctx,
		&itemRows,
		`SELECT order_id, product_id, count, total_price FROM order_items WHERE order_id = ?`,
		orderID,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	items := make([]model.OrderItem, len(itemRows))
	for i, row := range itemRows {
		items[i] = model.OrderItem{
			OrderID:   row.OrderID,
			ProductID: row.ProductID,
			Count:     row.Count,
			Price:     row.TotalPrice,
		}
	}

	return items, nil
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
