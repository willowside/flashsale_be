package repositoryiface

import "context"

type StockRepository interface {
	RestoreStock(ctx context.Context, productID string, qty int) error
}
