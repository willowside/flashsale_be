package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"flashsale/internal/cache"

	"github.com/redis/go-redis/v9"
)

type OrderProcessor struct {
	rdb *redis.Client
	lua *cache.LuaScripts
}

type OrderMessage struct {
	UserID    string `json:"user_id"`
	ProductID string `json:"product_id"`
	Timestamp int64  `json:"timestamp"`
}

func NewOrderProcessor(rdb *redis.Client, lua *cache.LuaScripts) *OrderProcessor {
	return &OrderProcessor{rdb: rdb, lua: lua}
}

func (p *OrderProcessor) ProcessOrder(ctx context.Context, body []byte) error {
	var msg OrderMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("invalid msg: %w", err)
	}

	userSetKey := fmt.Sprintf("flashsale:purchased:%s", msg.ProductID)

	// finalize via EvalSha
	_, err := p.rdb.EvalSha(ctx, p.lua.FinalizeSHA, []string{userSetKey}, msg.UserID).Result()
	if err != nil {
		return fmt.Errorf("finalize lua failed: %w", err)
	}

	// TODO: persist order to DB here if needed
	fmt.Println("Order Processed")
	return nil
}
