package worker

import (
	"context"
	"flashsale/internal/cache"
	"flashsale/internal/dto"
	"flashsale/internal/service"
	"fmt"
	"log"
)

type DLQWorker struct {
	compensator *service.OrderCompensator
}

func NewDLQWorker(c *service.OrderCompensator) *DLQWorker {
	return &DLQWorker{compensator: c}
}

func (w *DLQWorker) Handle(ctx context.Context, msg dto.DLQMessage) error {
	log.Printf("[DLQ worker] compensating order=%s reason=%s", msg.OrderNo, msg.Reason)
	// 1. mark DB order FAILED
	err := w.compensator.Compensate(ctx, msg)
	if err != nil {
		return fmt.Errorf("[DLQ worker] compensate order=%s: %w", msg.OrderNo, err)
	}
	// 2. update redis cache
	key := fmt.Sprintf("flashsale:order:%s", msg.OrderNo)
	cache.Rdb.HSet(ctx, key, "status", "FAILED")
	return nil
}
