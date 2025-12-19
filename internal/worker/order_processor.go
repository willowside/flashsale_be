package worker

import (
	"context"
	"encoding/json"
	"errors"
	"flashsale/internal/cache"
	"flashsale/internal/domain"
	"flashsale/internal/repository/repositoryiface"
	"flashsale/pkg/db"
	"fmt"
	"log"
)

type OrderProcessor struct {
	Repo       repositoryiface.OrderRepository
	LuaScripts *cache.LuaScripts
}

func NewOrderProcessor(repo repositoryiface.OrderRepository, lua *cache.LuaScripts) *OrderProcessor {
	return &OrderProcessor{Repo: repo, LuaScripts: lua}
}

func (p *OrderProcessor) ProcessOrder(ctx context.Context, body []byte) error {
	var msg domain.OrderMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Println("[worker] invalid msg:", err)
		return err
	}

	//ctx := context.Background()

	if p.LuaScripts == nil || p.LuaScripts.PrecheckSHA.SHA == "" {
		return errors.New("lua scripts not loaded")
	}

	// 1. lua finalize
	lua := p.LuaScripts.FinalizeSHA

	userSetKey := fmt.Sprintf("flashsale:purchased:%s", msg.ProductID)

	// finalize via EvalSha
	// _, err := p.rdb.EvalSha(ctx, p.lua.FinalizeSHA, []string{userSetKey}, msg.UserID).Result()

	allowed, err := lua.Run(ctx, cache.Rdb, []string{userSetKey}, msg.UserID).Bool()
	if err != nil {
		return fmt.Errorf("[worker] lua finalize failed: %w", err)
	}
	if !allowed {
		log.Println("[worker] order rejected: ", msg.UserID, msg.ProductID)
		return nil
	}

	// 2. DB distributed lock
	lockKey := fmt.Sprintf("lock:product:%s", &msg.ProductID)
	acquired, err := p.Repo.AcquireStockLock(ctx, lockKey, 5)
	if err != nil || !acquired {
		return fmt.Errorf("[worker] %s acquire lock failed: %w", msg.ProductID, err)
	}
	defer p.Repo.ReleaseStockLock(ctx, lockKey)

	// 3. DB transaction
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[worker] begin tx failed: %w", err)
	}
	defer tx.Rollback(ctx)

	// 3-1. reduce stock
	success, err := p.Repo.ReduceStock(ctx, msg.ProductID, 1)
	if err != nil {
		return fmt.Errorf("[worker] reduce stock failed: %w", err)

	}
	if !success {
		log.Println("[worker] empty stock: ", msg.ProductID)
		return nil
	}

	// 3-2. create order
	err = p.Repo.CreateOrder(ctx, msg.OrderID, msg.UserID, msg.ProductID)
	if err != nil {
		return fmt.Errorf("[worker] create order failed: %w", err)

	}

	// 4. commit
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[worker] commit failed: %w", err)

	}
	log.Println("[worker] order processed:", msg.UserID, msg.ProductID)
	return nil
}
