package usecase

import (
	"context"

	"kyc-service/internal/domain"
)

type GetClientUseCase struct {
	repo ApplicationRepository
}

func NewGetClientUseCase(repo ApplicationRepository) *GetClientUseCase {
	return &GetClientUseCase{repo: repo}
}

func (uc *GetClientUseCase) Execute(ctx context.Context, userId string) (*domain.Application, error) {
	app, err := uc.repo.GetByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, domain.ErrNotFound
	}

	return app, nil
}
