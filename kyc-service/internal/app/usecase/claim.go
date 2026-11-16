package usecase

import (
	"context"

	"kyc-service/internal/domain"
)

type ClaimRepository interface {
	ApplicationRepository
	Claim(ctx context.Context, id, officerID string) (bool, error)
}

type ClaimUseCase struct {
	repo ClaimRepository
}

func NewClaimUseCase(repo ClaimRepository) *ClaimUseCase {
	return &ClaimUseCase{repo: repo}
}

func (uc *ClaimUseCase) Execute(ctx context.Context, applicationID, officerID string) (*domain.Application, error) {
	// 1. Проверяем, что заявка существует и в статусе IN_REVIEW
	app, err := uc.repo.GetByID(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, domain.ErrNotFound
	}
	if app.Status != domain.StatusInReview {
		return nil, domain.ErrNotInReview
	}

	// 2. Оптимистическая блокировка
	ok, err := uc.repo.Claim(ctx, applicationID, officerID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrAlreadyClaimed
	}

	// 3. Возвращаем обновлённую заявку
	return uc.repo.GetByID(ctx, applicationID)
}
