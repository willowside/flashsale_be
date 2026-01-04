package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"flashsale/internal/domain"
	"flashsale/internal/repository/repositoryiface"

	"github.com/redis/go-redis/v9"
)

type FlashSaleRedisRepo struct {
	rdb *redis.Client
}

const ActiveFlashSaleKey = "flashsale:active:info"

func NewFlashSaleRedisRepo(rdb *redis.Client) repositoryiface.FlashSaleRedisRepository {
	return &FlashSaleRedisRepo{rdb: rdb}
}

func stockKey(productID int64) string {
	return fmt.Sprintf("flashsale:stock:%d", productID)
}

func (r *FlashSaleRedisRepo) WarmUpStock(
	ctx context.Context,
	flashSale domain.FlashSale,
	products []domain.FlashSaleProduct,
) error {

	ttl := flashSale.TTL(time.Now())
	if ttl <= 0 {
		return nil
	}

	// check to prevent overwrite by warmup
	// check if first product key exists to determine if was warmed
	if len(products) > 0 {
		firstKey := stockKey(products[0].ProductID)
		exists, err := r.rdb.Exists(ctx, firstKey).Result()
		if err != nil {
			return fmt.Errorf("check redis key existence failed: %w", err)
		}
		if exists > 0 {
			// skip warm up if data exists to avoid overwrite the stock deduction
			log.Printf("[warmup] stock already exists in redis, skipping to prevent override")
			return nil
		}
	}

	pipe := r.rdb.Pipeline()
	for _, p := range products {
		pipe.Set(
			ctx,
			stockKey(p.ProductID),
			p.SaleStock,
			ttl,
		)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *FlashSaleRedisRepo) GetStock(ctx context.Context, productID int64) (int64, error) {
	return r.rdb.Get(ctx, stockKey(productID)).Int64()
}

func (r *FlashSaleRedisRepo) SetActiveFlashSale(ctx context.Context, fs *domain.FlashSale, expiration time.Duration) error {
	data, err := json.Marshal(fs)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, ActiveFlashSaleKey, data, expiration).Err()
}

func (r *FlashSaleRedisRepo) GetActiveFlashSale(ctx context.Context) (*domain.FlashSale, error) {
	data, err := r.rdb.Get(ctx, ActiveFlashSaleKey).Bytes()
	if err == redis.Nil {
		return nil, nil // no cache
	} else if err != nil {
		return nil, err
	}

	var fs domain.FlashSale
	if err := json.Unmarshal(data, &fs); err != nil {
		return nil, err
	}
	return &fs, nil
}
