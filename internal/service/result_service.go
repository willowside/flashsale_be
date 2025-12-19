package service

import (
	"context"
	"flashsale/internal/application/dto"
	"flashsale/internal/repository/repositoryiface"
	"strconv"
)

type OrderResultService struct {
	repo repositoryiface.OrderRepository
}

func NewOrderResultService(repo repositoryiface.OrderRepository) *OrderResultService {
	return &OrderResultService{repo: repo}
}

func (s *OrderResultService) GetResult(
	ctx context.Context,
	orderID string,
) (*dto.OrderResult, error) {

	order, err := s.repo.GetByOrderID(ctx, orderID)
	if err != nil {
		// order not written yet → still in queue / worker processing
		return &dto.OrderResult{
			OrderUID:        orderID,
			ProcessingState: "PROCESSING",
		}, nil
	}

	// order exists → finalized
	return &dto.OrderResult{
		OrderUID:        strconv.FormatInt(order.ID, 10),
		Status:          order.Status,
		ProcessingState: "DONE",
		CreatedAt:       order.CreatedAt,
	}, nil
}
