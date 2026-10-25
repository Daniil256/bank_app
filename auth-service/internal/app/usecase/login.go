package usecase

import (
	"context"
	"fmt"
	"time"

	//"errors"
	"strings"

	"auth-service/internal/domain"
)

type LoginUseCase struct {
	repo        UserRepository
	signer      TokenSigner
	refreshRepo RefreshTokenRepository
	refreshTTL  time.Duration
}

type TokenSigner interface {
	SignAccess(userID, role string) (string, int64, error)
	//SignRefresh(userID string) (string, error)
}

func NewLoginUseCase(
	repo UserRepository,
	refreshRepo RefreshTokenRepository,
	signer TokenSigner,
	refreshTTL time.Duration,
) *LoginUseCase {
	return &LoginUseCase{
		repo:        repo,
		refreshRepo: refreshRepo,
		signer:      signer,
		refreshTTL:  refreshTTL,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, email, password string) (*domain.TokenPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		fmt.Println("[login] FindByEmail err ", err)
		return nil, err
	}
	fmt.Println("[login] User ", user)
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !verifyPassword(password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}
	// Подписываем access
	accessToken, expiresIn, err := uc.signer.SignAccess(user.ID, user.Role)
	if err != nil {
		fmt.Println("[login] SignAccess err ", err)
		return nil, err
	}

	// Refresh — случайная строка + хеш в БД
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(uc.refreshTTL)

	if err := uc.refreshRepo.Create(ctx, user.ID, hashRefreshToken(refreshToken), expiresAt); err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}
