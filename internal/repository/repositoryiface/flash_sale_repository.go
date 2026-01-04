package repositoryiface

import (
	"context"
	"flashsale/internal/domain"
	"time"
)

type FlashSaleRepository interface {
	// DB
	GetActiveFlashSale(ctx context.Context) (*domain.FlashSale, error)
	GetFlashSaleByID(ctx context.Context, id int64) (*domain.FlashSale, error)
	GetFlashSaleProducts(ctx context.Context, flashSaleID int64) ([]domain.FlashSaleProduct, error)
	GetFlashSaleProduct(ctx context.Context, flashSaleID int64, productID string) (*domain.FlashSaleProduct, error)
	UpdateStatus(ctx context.Context, id int64, status domain.FlashSaleStatus) error
}

type FlashSaleRedisRepository interface {
	// Redis
	WarmUpStock(
		ctx context.Context,
		flashSale domain.FlashSale,
		products []domain.FlashSaleProduct,
	) error

	GetStock(ctx context.Context, productID int64) (int64, error)

	SetActiveFlashSale(ctx context.Context, fs *domain.FlashSale, expiration time.Duration) error
	GetActiveFlashSale(ctx context.Context) (*domain.FlashSale, error)
}
