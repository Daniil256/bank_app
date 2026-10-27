package usecase

import (
	"context"

	"customer-service/internal/domain"
)

type GetStatusUseCase struct {
	repo CustomerRepository
}

func NewGetStatusUseCase(repo CustomerRepository) *GetStatusUseCase {
	return &GetStatusUseCase{repo: repo}
}

func (uc *GetStatusUseCase) Execute(ctx context.Context, customerID string) (*domain.Customer, error) {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, domain.ErrNotFound
	}
	return customer, nil
}
