package repositoryiface

import (
	"context"
	"flashsale/internal/domain"
)

type OrderRepository interface {
	// create order record
	CreateOrder(ctx context.Context, orderID, userID, productID string) error
	GetByOrderID(ctx context.Context, orderID string) (*domain.Order, error)

	// decrease stock, return true if stock > 0 -> reduce success
	ReduceStock(ctx context.Context, productID string, qty int64) (bool, error)

	// check remaining stock
	GetStock(ctx context.Context, productID string) (int64, error)

	// distributed lock for stock reduction
	AcquireStockLock(ctx context.Context, key string, ttlSeconds int) (bool, error)
	ReleaseStockLock(ctx context.Context, key string) error
}
