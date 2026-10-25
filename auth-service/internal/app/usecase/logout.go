package usecase

import (
	"auth-service/internal/domain"
	"context"
	"fmt"
)

type LogoutUseCase struct {
	refreshRepo RefreshTokenRepository
}

func NewLogoutUseCase(refreshRepo RefreshTokenRepository) *LogoutUseCase {
	return &LogoutUseCase{refreshRepo: refreshRepo}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, refreshToken string) error {
	fmt.Println("[logout.Execute] refreshToken ", refreshToken)
	hash := hashRefreshToken(refreshToken)

	err := uc.refreshRepo.Revoke(ctx, hash)
	if err != nil {
		return domain.ErrInvalidLogout
	}
	return nil
}
