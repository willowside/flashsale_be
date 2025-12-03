package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client
var Ctx = context.Background()

func InitRedis(host, port, password string, db int) error {
	addr := fmt.Sprintf("%s:%s", host, port)

	Rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		PoolSize: 20,
	})

	_, err := Rdb.Ping(Ctx).Result()
	if err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}
