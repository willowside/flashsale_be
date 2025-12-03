package repository

import (
	"context"
	"flashsale/internal/domain/flashsale"
	"flashsale/internal/domain/order"
	"flashsale/internal/domain/product"
)

type ProductRepository interface {
	GetByID(ctx context.Context, id int64) (*product.Product, error)
	DecreaseStock(ctx context.Context, productID int64) error
}

type FlashSaleRepository interface {
	GetActiveSale(ctx context.Context, productID int64) (*flashsale.FlashSale, error)
}

type OrderRepository interface {
	CreateTx(ctx context.Context, tx Tx, o *order.Order) error
}

type Tx interface {
	Exec(ctx context.Context, query string, args ...any) error
	QueryRow(ctx context.Context, query string, args ...any) Row
	Commit() error
	Rollback() error
}

type Row interface {
	Scan(des ...any) error
}

type DB interface {
	BeginTx(ctx context.Context) (Tx, error)
}
