package usecase

import (
	"context"

	"bank_app/shared/roles"
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

func (uc *GetCustomerUseCase) Execute(ctx context.Context, customerID, requesterRole, requesterCustomerID string) (*domain.Customer, error) {
	// CLIENT — только себя
	if requesterRole == roles.Client && customerID != requesterCustomerID {
		return nil, domain.ErrAccessDenied
	}
	// OFFICER/ADMIN — любого
	if requesterRole != roles.Client && requesterRole != roles.Officer && requesterRole != roles.Admin {
		return nil, domain.ErrAccessDenied
	}

	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, domain.ErrNotFound
	}
	return customer, nil
}
