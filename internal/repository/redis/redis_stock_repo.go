package redis

import (
	"context"
	"flashsale/internal/repository/repositoryiface"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisStockRepository struct {
	rdb *redis.Client
}

func NewRedisStockRepo(rdb *redis.Client) repositoryiface.RedisStockRepository {
	return &RedisStockRepository{
		rdb: rdb,
	}
}

func (r *RedisStockRepository) RestoreStock(ctx context.Context, productID string, qty int) error {
	key := fmt.Sprintf("flashsale:stock:%s", productID)
	return r.rdb.Incr(ctx, key).Err()
}
