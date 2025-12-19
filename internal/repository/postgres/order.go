package postgres

import (
	"context"
	"flashsale/internal/cache"
	"flashsale/internal/domain"
	"flashsale/internal/repository/repositoryiface"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// concrete implementation of order repository for Postgres
// 1. New()
// 2. implement methods

// 💡 Struct updated to hold the Locker dependency
type OrderPGRepo struct {
	Pool *pgxpool.Pool
}

// 💡 Constructor now accepts the Locker interface
func NewOrderPGRepo(pool *pgxpool.Pool) repositoryiface.OrderRepository {
	return &OrderPGRepo{
		Pool: pool,
	}
}

// insert row
func (r *OrderPGRepo) CreateOrder(ctx context.Context, orderID, userID, productID string) error {
	_, err := r.Pool.Exec(ctx,
		`INSERT INTO orders(id, user_id, product_id, flash_sale_id, price, status, created_at)
		VALUES ($4, $1, $2, 1, 1280, $3, NOW())`, userID, productID, "", orderID)
	return err
}

func (r *OrderPGRepo) GetByOrderID(
	ctx context.Context,
	orderID string,
) (*domain.Order, error) {

	row := r.Pool.QueryRow(ctx, `
		SELECT
			id,
			user_id,
			product_id,
			flash_sale_id,
			price,
			status,
			created_at
		FROM orders
		WHERE id = $1`, orderID)

	var o domain.Order
	err := row.Scan(
		&o.ID,
		&o.UserID,
		&o.ProductID,
		&o.FlashSaleID,
		&o.Price,
		&o.Status,
		&o.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &o, nil
}

// UPDATE: reduce stock
func (r *OrderPGRepo) ReduceStock(ctx context.Context, productID string, qty int64) (bool, error) {

	res, err := r.Pool.Exec(ctx,
		`UPDATE products SET stock = stock - $2 WHERE id = $1 AND stock >= $2`,
		productID, qty)
	if err != nil {
		return false, err
	}

	if res.RowsAffected() == 0 {
		return false, nil
	}
	return true, nil
}

func (r *OrderPGRepo) GetStock(ctx context.Context, productID string) (int64, error) {
	var stock int64
	err := r.Pool.QueryRow(ctx,
		`SELECT products WHERE id = $1`, productID).Scan(&stock)
	return stock, err
}

// Redis lock, SET NX
func (r *OrderPGRepo) AcquireStockLock(ctx context.Context, key string, ttlSeconds int) (bool, error) {
	ok, err := cache.Rdb.SetNX(ctx, key, "1", time.Duration(ttlSeconds)*time.Second).Result()
	return ok, err
}

func (r *OrderPGRepo) ReleaseStockLock(ctx context.Context, key string) error {
	_, err := cache.Rdb.Del(ctx, key).Result()
	return err
}
