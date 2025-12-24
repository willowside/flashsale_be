package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisStockRepository struct {
	rdb *redis.Client
}

func (r *RedisStockRepository) RestoreStock(ctx context.Context, productID string, qty int) error {
	key := fmt.Sprintf("flashsale:stock:%s", productID)
	return r.rdb.IncrBy(ctx, key, int64(qty)).Err()
}
