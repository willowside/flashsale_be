package repositoryiface

import "context"

type ProductRepository interface {
	GetStock(ctx context.Context, productID int64) (int64, error)
}
