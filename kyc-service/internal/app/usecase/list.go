package usecase

import (
	"context"

	"kyc-service/internal/domain"
)

type ApplicationListRepository interface {
	ApplicationRepository
	List(ctx context.Context, status string, limit int, cursor string) ([]domain.Application, string, error)
}

type ListUseCase struct {
	repo ApplicationListRepository
}

func NewListUseCase(repo ApplicationListRepository) *ListUseCase {
	return &ListUseCase{repo: repo}
}

func (uc *ListUseCase) Execute(ctx context.Context, requesterRole, status string, limit int, cursor string) ([]domain.Application, string, error) {
	// Только OFFICER/ADMIN
	if requesterRole != "OFFICER" && requesterRole != "ADMIN" {
		return nil, "", domain.ErrAccessDenied
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return uc.repo.List(ctx, status, limit, cursor)
}
