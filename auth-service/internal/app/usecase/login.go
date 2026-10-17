package usecase

import (
	"context"
	//"errors"
	"strings"

	"auth-service/internal/domain"
)

type LoginUseCase struct {
	repo UserRepository
}

func NewLoginUseCase(repo UserRepository) *LoginUseCase {
	return &LoginUseCase{repo: repo}
}

func (uc *LoginUseCase) Execute(ctx context.Context, email, password string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !verifyPassword(password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	return user, nil
}
