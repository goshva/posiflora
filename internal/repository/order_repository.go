package repository

import (
	"context"

	"posiflora-mvp/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

func (r *OrderRepository) Upsert(ctx context.Context, shopID int64, payload model.OrderPayload) (*model.Order, bool, error) {
	var order model.Order
	created := false

	query := `
	INSERT INTO orders (shop_id, number, total, customer_name, created_at)
	VALUES ($1, $2, $3, $4, NOW())
	ON CONFLICT (shop_id, number) DO NOTHING
	RETURNING id, number, total, customer_name, created_at
	`

	err := r.pool.QueryRow(ctx, query, shopID, payload.Number, payload.Total, payload.CustomerName).
		Scan(&order.ID, &order.Number, &order.Total, &order.CustomerName, &order.CreatedAt)

	if err == pgx.ErrNoRows {
		// Уже существует
		err = r.pool.QueryRow(ctx, `
			SELECT id, number, total, customer_name, created_at
			FROM orders
			WHERE shop_id = $1 AND number = $2
		`, shopID, payload.Number).Scan(&order.ID, &order.Number, &order.Total, &order.CustomerName, &order.CreatedAt)
		if err != nil {
			return nil, false, err
		}
	} else if err == nil {
		created = true
	} else {
		return nil, false, err
	}

	order.ShopID = shopID
	return &order, created, nil
}

// LogSend записывает результат отправки в Telegram
func (r *OrderRepository) LogSend(ctx context.Context, shopID, orderID int64, message, status string, errText *string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO telegram_send_log (shop_id, order_id, message, status, error)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (shop_id, order_id) DO NOTHING
	`, shopID, orderID, message, status, errText)
	return err
}

// GetSendStatus возвращает статус отправки
func (r *OrderRepository) GetSendStatus(ctx context.Context, shopID, orderID int64) (map[string]interface{}, error) {
	var status string
	err := r.pool.QueryRow(ctx, `
		SELECT status FROM telegram_send_log
		WHERE shop_id=$1 AND order_id=$2
	`, shopID, orderID).Scan(&status)
	if err != nil {
		status = "skipped"
	}
	return map[string]interface{}{"sendStatus": status}, nil
}

func (r *OrderRepository) HasSendLog(ctx context.Context, shopID, orderID int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM telegram_send_log
			WHERE shop_id=$1 AND order_id=$2
		)
	`, shopID, orderID).Scan(&exists)
	return exists, err
}
