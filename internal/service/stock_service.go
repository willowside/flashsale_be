package service

import (
	"context"
	"flashsale/internal/repository/repositoryiface"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type StockService interface {
	GetStock(ctx context.Context, productID int64) (int64, error)
}

type stockerService struct {
	rdb         *redis.Client
	productRepo repositoryiface.ProductRepository
}

func NewStockService(rdb *redis.Client, productRepo repositoryiface.ProductRepository) StockService {
	return &stockerService{
		rdb:         rdb,
		productRepo: productRepo,
	}
}

func (s *stockerService) GetStock(ctx context.Context, productID int64) (int64, error) {
	key := fmt.Sprintf("flashsale:stock:%d", productID)

	// 1. check Redis first (fast path)
	if v, _ := s.rdb.Get(ctx, key).Int64(); v >= 0 {
		return v, nil
	}

	// 2. DB fallback
	stock, err := s.productRepo.GetStock(ctx, productID)
	if err != err {
		return 0, err
	}

	// 3. update redis
	s.rdb.Set(ctx, key, stock, 60) // 60s cache

	return stock, nil

}
