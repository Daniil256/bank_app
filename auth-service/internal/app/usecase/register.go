package usecase

import (
	"context"
	"crypto/rand"
	"strings"

	"auth-service/internal/domain"
	"golang.org/x/crypto/argon2"
)

type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
}

type EventPublisher interface {
	PublishUserRegistered(userID, email string) error
}

type RegisterUseCase struct {
	repo   UserRepository
	events EventPublisher
}

func NewRegisterUseCase(repo UserRepository, events EventPublisher) *RegisterUseCase {
	return &RegisterUseCase{repo: repo, events: events}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, email, password string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	if !strings.Contains(email, "@") {
		return "", domain.ErrInvalidEmail
	}
	if len(password) < 8 {
		return "", domain.ErrWeakPassword
	}

	exists, err := uc.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if exists {
		return "", domain.ErrEmailTaken
	}

	hash, err := hashPassword(password)
	if err != nil {
		return "", err
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: hash,
		Role:         domain.RoleClient,
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		return "", err
	}

	_ = uc.events.PublishUserRegistered(user.ID, user.Email)

	return user.ID, nil
}

// hashPassword — argon2id с дефолтными параметрами
func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return EncodeArgon2(salt, hash), nil
}
