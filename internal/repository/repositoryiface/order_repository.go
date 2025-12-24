package repositoryiface

import (
	"context"
	"flashsale/internal/domain"

	"github.com/jackc/pgx/v5"
)

type OrderRepository interface {
	CreatePendingOrder(ctx context.Context, orderNo, userID, productID string) error
	GetOrderStatus(ctx context.Context, orderNo string) (string, error)

	BeginTx(ctx context.Context) (pgx.Tx, error)

	ReduceStockTx(ctx context.Context, tx pgx.Tx, productID string, qty int64) (bool, error)
	MarkOrderSuccessTx(ctx context.Context, tx pgx.Tx, orderNo string) error
	MarkOrderFailedTx(ctx context.Context, tx pgx.Tx, orderNo string, reason string) error

	GetByOrderNo(ctx context.Context, orderID string) (*domain.Order, error)

	// decrease stock, return true if stock > 0 -> reduce success
	// ReduceStock(ctx context.Context, productID string, qty int64) (bool, error)

	// check remaining stock
	GetStock(ctx context.Context, productID string) (int64, error)

	// distributed lock for stock reduction
	AcquireStockLock(ctx context.Context, key string, ttlSeconds int) (bool, error)
	ReleaseStockLock(ctx context.Context, key string) error

	// MarkOrderSuccess(ctx context.Context, orderNo string) error
	MarkOrderFailed(ctx context.Context, orderNo string, reason string) error
}
