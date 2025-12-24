package service

import (
	"context"
	"flashsale/internal/dto"
	"flashsale/internal/repository/repositoryiface"
)

type OrderResultService struct {
	repo repositoryiface.OrderRepository
}

func NewOrderResultService(repo repositoryiface.OrderRepository) *OrderResultService {
	return &OrderResultService{repo: repo}
}

func (s *OrderResultService) GetResult(ctx context.Context, orderID string) (*dto.OrderResult, error) {
	status, err := s.repo.GetOrderStatus(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return &dto.OrderResult{
		OrderID: orderID,
		Status:  status,
	}, nil
}
