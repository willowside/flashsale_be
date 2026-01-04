package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"flashsale/internal/domain"
	"flashsale/internal/repository/repositoryiface"
)

type FlashSaleWarmUpService struct {
	dbRepo    repositoryiface.FlashSaleRepository
	redisRepo repositoryiface.FlashSaleRedisRepository
}

func NewFlashSaleWarmUpService(
	dbRepo repositoryiface.FlashSaleRepository,
	redisRepo repositoryiface.FlashSaleRedisRepository,
) *FlashSaleWarmUpService {
	return &FlashSaleWarmUpService{
		dbRepo:    dbRepo,
		redisRepo: redisRepo,
	}
}

func (s *FlashSaleWarmUpService) WarmUpByID(ctx context.Context, id int64) error {
	// 1. get flashsale info
	fs, err := s.dbRepo.GetFlashSaleByID(ctx, id)
	if err != nil {
		return err
	}

	// if flashsale isActive, run Redis warm-up
	if fs.Status == domain.StatusEnded {
		return fmt.Errorf("sale ended, cannot warm-up")
	}

	products, err := s.dbRepo.GetFlashSaleProducts(ctx, fs.ID)
	if err != nil {
		return err
	}

	// run redis warmup(Exists checking within)
	if err := s.redisRepo.WarmUpStock(ctx, *fs, products); err != nil {
		return fmt.Errorf("[warmup] Redis warm-up failed: %w", err)
	}

	ttl := fs.TTL(time.Now())
	if err := s.redisRepo.SetActiveFlashSale(ctx, fs, ttl); err != nil {
		// log & no interrupt process, warm-up is successful
		log.Printf("[warmup] warn: flashsale info cache update failed: %v", err)
	} else {
		log.Printf("[warmup] flashsale info synced to redis cache")
	}

	// update flashsale status active in DB
	// only update status when it's "scheduled", to avoid duplicated triggering
	if fs.Status == domain.StatusScheduled {
		if err := s.dbRepo.UpdateStatus(ctx, id, domain.StatusActive); err != nil {
			return fmt.Errorf("[warmup] Flashsale update failed: %w", err)
		}
		log.Printf("[warmup] flashsale %d status activated", id)
	}

	return nil
}

func (s *FlashSaleWarmUpService) WarmUp(ctx context.Context) error {
	fs, err := s.dbRepo.GetActiveFlashSale(ctx)
	if err != nil {
		return err
	}

	if !fs.IsActive(time.Now()) {
		log.Println("[warmup] flash sale not in active window")
		return nil
	}

	products, err := s.dbRepo.GetFlashSaleProducts(ctx, fs.ID)
	if err != nil {
		return err
	}

	// Pass the pointer if necessary, or dereference as you did
	if err := s.redisRepo.WarmUpStock(ctx, *fs, products); err != nil {
		return err
	}

	log.Printf("[warmup] success flash_sale=%d products=%d", fs.ID, len(products))
	return nil
}
