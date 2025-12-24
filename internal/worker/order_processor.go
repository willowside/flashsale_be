package worker

import (
	"context"
	"encoding/json"
	"errors"
	"flashsale/internal/cache"
	"flashsale/internal/dto"
	"flashsale/internal/repository/repositoryiface"
	"fmt"
)

var (
	ErrOutOfStock = errors.New("out_of_stock")
	ErrLuaReject  = errors.New("lua_reject")
)

type OrderProcessor struct {
	Repo       repositoryiface.OrderRepository
	LuaScripts *cache.LuaScripts
}

func NewOrderProcessor(repo repositoryiface.OrderRepository, lua *cache.LuaScripts) *OrderProcessor {
	return &OrderProcessor{Repo: repo, LuaScripts: lua}
}

// 1. deal ONE order
// 2. return err -> worker decides retry/DLQ

func (p *OrderProcessor) ProcessOrder(ctx context.Context, body []byte) error {
	var msg dto.OrderMessage

	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}

	// 0. Idempotency check
	status, err := p.Repo.GetOrderStatus(ctx, msg.OrderID)
	if err != nil {
		return err
	}
	if status == "SUCCESS" || status == "FAILED" {
		return nil // already processed
	}

	if p.LuaScripts == nil || p.LuaScripts.PrecheckSHA.SHA == "" {
		return errors.New("lua scripts not loaded")
	}

	// 1. redis lua finalize (prevent duplicated purchasing)
	lua := p.LuaScripts.FinalizeSHA
	userSetKey := fmt.Sprintf("flashsale:purchased:%s", msg.ProductID)
	allowed, err := lua.Run(ctx, cache.Rdb, []string{userSetKey}, msg.UserID).Bool()
	if err != nil {
		return fmt.Errorf("[worker] lua finalize failed: %w", err)
	}
	if !allowed {
		return ErrLuaReject
	}

	// 2. DB distributed lock
	lockKey := fmt.Sprintf("lock:product:%s", msg.ProductID)
	acquired, err := p.Repo.AcquireStockLock(ctx, lockKey, 5)
	if err != nil || !acquired {
		return fmt.Errorf("[worker] %s acquire lock failed: %w", msg.ProductID, err)
	}
	defer p.Repo.ReleaseStockLock(ctx, lockKey)

	// 3. DB transaction
	tx, err := p.Repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("[worker] begin tx failed: %w", err)
	}
	defer tx.Rollback(ctx)

	// 3-1. reduce stock
	success, err := p.Repo.ReduceStockTx(ctx, tx, msg.ProductID, 1)
	if err != nil {
		return fmt.Errorf("[worker] reduce stock failed: %w", err)
	}
	if !success {
		_ = p.Repo.MarkOrderFailedTx(ctx, tx, msg.OrderID, "OUT_OF_STOCK")
		_ = tx.Commit(ctx)
		return ErrOutOfStock
	}

	// 3-2. Mark success
	err = p.Repo.MarkOrderSuccessTx(ctx, tx, msg.OrderID)
	if err != nil {
		return fmt.Errorf("[worker] create order failed: %w", err)

	}
	return tx.Commit(ctx)
}
