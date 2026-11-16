package usecase

import (
	"context"

	"customer-service/internal/domain"
)

type CustomerRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Customer, error)
	Update(ctx context.Context, c *domain.Customer) error
}

type GetCustomerUseCase struct {
	repo CustomerRepository
}

func NewGetCustomerUseCase(repo CustomerRepository) *GetCustomerUseCase {
	return &GetCustomerUseCase{repo: repo}
}

func (uc *GetCustomerUseCase) Execute(ctx context.Context, userID string) (*domain.Customer, error) {
	customer, err := uc.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, domain.ErrNotFound
	}
	return customer, nil
}
