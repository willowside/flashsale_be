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

func (w *DLQWorker) Handle(ctx context.Context, msg dto.DLQMessage) {
	log.Printf("[DLQ] compensating order=%s reason=%s", msg.OrderNo, msg.Reason)
	// 1. mark DB order FAILED
	w.compensator.Compensate(ctx, msg)
	// 2. update redis cache
	key := fmt.Sprintf("falshsale:order:%s", msg.OrderNo)
	cache.Rdb.HSet(ctx, key, "status", "FAILED")
}
