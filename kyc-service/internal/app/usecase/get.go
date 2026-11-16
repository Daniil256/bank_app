package usecase

import (
	"context"

	"kyc-service/internal/domain"
)

type GetUseCase struct {
	repo ApplicationRepository
}

func NewGetUseCase(repo ApplicationRepository) *GetUseCase {
	return &GetUseCase{repo: repo}
}

func (uc *GetUseCase) Execute(ctx context.Context, applicationID string) (*domain.Application, error) {
	app, err := uc.repo.GetByID(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, domain.ErrNotFound
	}

	return app, nil
}
