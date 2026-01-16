package repository

import (
	"context"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"product/pkg/product/domain/model"
)

func NewReservationRepository(ctx context.Context, client mysql.ClientContext) model.ReservationRepository {
	return &reservationRepository{
		ctx:    ctx,
		client: client,
	}
}

type reservationRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (r *reservationRepository) Store(reservation *model.ProductReservation) error {
	_, err := r.client.ExecContext(r.ctx,
		`
		INSERT INTO product_reservation (order_id, product_id, count, price, created_at) 
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			count=VALUES(count),
			price=VALUES(price),
			created_at=VALUES(created_at)
		`,
		reservation.OrderID,
		reservation.ProductID,
		reservation.Count,
		reservation.Price,
		reservation.CreatedAt,
	)
	return errors.WithStack(err)
}

func (r *reservationRepository) FindKeysByOrderID(orderID uuid.UUID) ([]model.ProductReservation, error) {
	var rows []struct {
		OrderID   uuid.UUID `db:"order_id"`
		ProductID uuid.UUID `db:"product_id"`
		Count     int       `db:"count"`
		Price     float64   `db:"price"`
		CreatedAt time.Time `db:"created_at"`
	}

	err := r.client.SelectContext(
		r.ctx,
		&rows,
		`SELECT order_id, product_id, count, price, created_at 
		 FROM product_reservation 
		 WHERE order_id = ?`,
		orderID,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := make([]model.ProductReservation, len(rows))
	for i, row := range rows {
		result[i] = model.ProductReservation{
			OrderID:   row.OrderID,
			ProductID: row.ProductID,
			Count:     row.Count,
			Price:     row.Price,
			CreatedAt: row.CreatedAt,
		}
	}

	return result, nil
}

func (r *reservationRepository) RemoveByOrderID(orderID uuid.UUID) error {
	_, err := r.client.ExecContext(r.ctx,
		`DELETE FROM product_reservation WHERE order_id = ?`,
		orderID,
	)
	return errors.WithStack(err)
}
