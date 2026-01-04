package repositoryiface

import "context"

type StockRepository interface {
	GetStock(ctx context.Context, productID int64) (int64, error)
}

type RedisStockRepository interface {
	RestoreStock(ctx context.Context, productID string, qty int) error
}
