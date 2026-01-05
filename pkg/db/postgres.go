package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func InitPostgresDB(host, port, user, password, db_name, ssl_mode string) error {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, db_name, ssl_mode)
	var err error
	var pool *pgxpool.Pool

	// 💡 嘗試連線多次，總共等待約 30 秒
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		pool, err = pgxpool.New(ctx, dsn)
		if err == nil {
			err = pool.Ping(ctx)
			if err == nil {
				cancel()
				Pool = pool
				return nil
			}
		}

		cancel()
		fmt.Printf("等待 Postgres 啟動中 (%d/10): %v\n", i+1, err)
		time.Sleep(3 * time.Second) // 每 3 秒重試一次
	}

	return fmt.Errorf("無法連線至 Postgres，已達最大重試次數: %w", err)
}
