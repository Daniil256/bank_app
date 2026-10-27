package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"auth-service/internal/domain"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	FindByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, tokenHash string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

type UserFinder interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)
}

type RefreshUseCase struct {
	refreshRepo RefreshTokenRepository
	userRepo    UserFinder
	signer      TokenSigner
	refreshTTL  time.Duration
}

func NewRefreshUseCase(
	refreshRepo RefreshTokenRepository,
	userRepo UserFinder,
	signer TokenSigner,
	refreshTTL time.Duration,
) *RefreshUseCase {
	return &RefreshUseCase{
		refreshRepo: refreshRepo,
		userRepo:    userRepo,
		signer:      signer,
		refreshTTL:  refreshTTL,
	}
}

func (uc *RefreshUseCase) Execute(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	fmt.Println("[refresh.Execute] refreshToken ", refreshToken)
	// 1. Хешируем входящий refresh
	hash := hashRefreshToken(refreshToken)

	// 2. Ищем в БД
	stored, err := uc.refreshRepo.FindByHash(ctx, hash)
	if err != nil {
		fmt.Println("[refresh.Execute] FindByHash error ", err)
		return nil, err
	}
	fmt.Println("[refresh.Execute] FindByHash token ", stored)
	if stored == nil {
		return nil, nil
	}

	// 3. Проверяем активность
	if !stored.IsActive() {
		fmt.Println("[refresh.Execute] token is not IsActive, revoke ", stored.UserID)
		// ⚠️ Если токен отозван, но кто-то его использует —
		// это сигнал компрометации. Отзываем ВСЕ токены пользователя.
		if stored.RevokedAt != nil {
			_ = uc.refreshRepo.RevokeAllForUser(ctx, stored.UserID)
		}
		return nil, err
	}

	// 4. Загружаем пользователя (нужны role, customer_id)
	user, err := uc.userRepo.FindByID(ctx, stored.UserID)
	if err != nil {
		return nil, fmt.Errorf("[refresh.Execute] FindByID error %v", err)
	}
	fmt.Println("[refresh.Execute] FindByID user ", user)

	if user == nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	// 5. Ротация: отзываем старый refresh
	if err := uc.refreshRepo.Revoke(ctx, hash); err != nil {
		fmt.Println("[refresh.Execute] Revoke error ", err)
		return nil, err
	}

	// 6. Генерируем новый refresh (случайная строка)
	newRefresh, err := generateRefreshToken()
	if err != nil {
		fmt.Println("[refresh.Execute] generateRefreshToken error ", err)
		return nil, err
	}
	newHash := hashRefreshToken(newRefresh)
	expiresAt := time.Now().Add(uc.refreshTTL)

	if err := uc.refreshRepo.Create(ctx, user.ID, newHash, expiresAt); err != nil {
		fmt.Println("[refresh.Execute] Create error ", err)
		return nil, err
	}

	// 7. Подписываем новый access
	accessToken, expiresIn, err := uc.signer.SignAccess(user.ID, user.Role, user.CustomerID)
	if err != nil {
		fmt.Println("[refresh.Execute] SignAccess error ", err)
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefresh,
		ExpiresIn:    expiresIn,
	}, nil
}

// generateRefreshToken — 32 байта случайности → hex
func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		fmt.Println("generate refresh token error:", err)
		return "", err
	}
	return hex.EncodeToString(b), nil
}
