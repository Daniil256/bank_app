package usecase

import (
	"context"
	"fmt"
	"strings"

	"auth-service/internal/domain"
)

type CreateOfficerUseCase struct {
	repo UserRepository
}

func NewCreateOfficerUseCase(repo UserRepository) *CreateOfficerUseCase {
	return &CreateOfficerUseCase{repo: repo}
}

func (uc *CreateOfficerUseCase) Execute(ctx context.Context, email, password string) (string, error) {
	fmt.Println("[create_officer.Execute] email ", email)
	email = strings.ToLower(strings.TrimSpace(email))

	if !strings.Contains(email, "@") {
		return "", domain.ErrInvalidEmail
	}
	if len(password) < 8 {
		return "", domain.ErrWeakPassword
	}

	exists, err := uc.repo.ExistsByEmail(ctx, email)

	if err != nil {
		fmt.Println("[create_officer.ExistsByEmail] err ", err)
		return "", err
	}

	fmt.Println("[create_officer.ExistsByEmail] exists ", exists)
	if exists {
		return "", domain.ErrEmailExists
	}

	hash, err := hashPassword(password)
	if err != nil {
		fmt.Println("[create_officer.hashPassword] err ", err)
		return "", err
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: hash,
		Role:         domain.RoleOfficer,
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		fmt.Println("[create_officer.Create] err ", err)
		return "", err
	}

	return user.ID, nil
}
