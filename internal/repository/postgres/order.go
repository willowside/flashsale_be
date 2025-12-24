package postgres

import (
	"context"
	"flashsale/internal/cache"
	"flashsale/internal/domain"
	"flashsale/internal/repository/repositoryiface"
	"time"

	"github.com/jackc/pgx/v5"
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

func (r *OrderPGRepo) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.Pool.Begin(ctx)
}

func (r *OrderPGRepo) CreatePendingOrder(ctx context.Context, orderNo, userID, productID string) error {
	_, err := r.Pool.Exec(ctx, `
		INSERT INTO orders (order_no, user_id, product_id, status)
		VALUES ($1, $2, $3, 'pending')
		ON CONFLICT (order_no) DO NOTHING;
	`, orderNo, userID, productID)
	return err
}

func (r *OrderPGRepo) GetByOrderNo(ctx context.Context, orderNo string) (*domain.Order, error) {
	row := r.Pool.QueryRow(ctx, `
		SELECT id, order_no, user_id, product_id, flash_sale_id,
		       price, status, created_at, paid_at, canceled_at
		FROM orders
		WHERE order_no = $1
	`, orderNo)

	var o domain.Order
	err := row.Scan(
		&o.ID,
		&o.OrderNo,
		&o.UserID,
		&o.ProductID,
		&o.FlashSaleID,
		&o.Price,
		&o.Status,
		&o.CreatedAt,
		&o.PaidAt,
		&o.CanceledAt,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// UPDATE: reduce stock
func (r *OrderPGRepo) ReduceStockTx(ctx context.Context, tx pgx.Tx, productID string, qty int64) (bool, error) {

	res, err := tx.Exec(ctx,
		`UPDATE products SET stock = stock - $2 WHERE id = $1 AND stock >= $2`,
		productID, qty)
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, nil
}

func (r *OrderPGRepo) GetStock(ctx context.Context, productID string) (int64, error) {
	var stock int64
	err := r.Pool.QueryRow(ctx,
		`SELECT stock FROM products WHERE id = $1`, productID).Scan(&stock)
	return stock, err
}

func (r *OrderPGRepo) GetOrderStatus(ctx context.Context, orderNo string) (string, error) {
	var status string
	err := r.Pool.QueryRow(ctx,
		`SELECT status FROM orders WHERE order_no = $1`, orderNo).Scan(&status)
	return status, err
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

func (r *OrderPGRepo) MarkOrderSuccessTx(ctx context.Context, tx pgx.Tx, orderNo string) error {
	_, err := tx.Exec(ctx, `
		UPDATE orders
		SET status = 'success', paid_at = NOW()
		WHERE order_no = $1 AND status = 'pending'
	`, orderNo)
	return err
}

func (r *OrderPGRepo) MarkOrderFailedTx(ctx context.Context, tx pgx.Tx, orderNo string, reason string) error {
	_, err := tx.Exec(ctx, `
		UPDATE orders
		SET status = 'failed', canceled_at = NOW()
		WHERE order_no = $1 AND status = 'pending'
	`, orderNo)
	return err
}

func (r *OrderPGRepo) MarkOrderFailed(ctx context.Context, orderNo string, reason string) error {
	_, err := r.Pool.Exec(ctx, `
		UPDATE orders
		SET status = 'failed', canceled_at = NOW()
		WHERE order_no = $1
		  AND status = 'pending'
	`, orderNo)
	return err
}
