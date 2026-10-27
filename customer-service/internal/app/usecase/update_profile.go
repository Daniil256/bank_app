package usecase

import (
	"context"
	"time"

	"bank_app/shared/roles"
	"customer-service/internal/domain"
)

type EventPublisher interface {
	PublishProfileFilled(customerID, userID string) error
}

type UpdateProfileUseCase struct {
	repo   CustomerRepository
	events EventPublisher
}

func NewUpdateProfileUseCase(repo CustomerRepository, events EventPublisher) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{repo: repo, events: events}
}

type UpdateProfileInput struct {
	FirstName   string
	LastName    string
	BirthDate   string // YYYY-MM-DD
	Address     string
	Phone       string
	Citizenship string
}

func (uc *UpdateProfileUseCase) Execute(ctx context.Context, customerID string, in UpdateProfileInput) (*domain.Customer, error) {
	if customerID == "" {
		return nil, domain.ErrCustomerIDRequired
	}

	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, domain.ErrNotFound
	}

	// Нельзя редактировать после ON_KYC
	if customer.Status == roles.StatusOnKyc || customer.Status == roles.StatusActive {
		return nil, domain.ErrCannotEditOnKyc
	}

	// Валидация
	if in.FirstName == "" || in.LastName == "" {
		return nil, domain.ErrInvalidProfile
	}

	// Парсим дату рождения
	var birthDate *time.Time
	if in.BirthDate != "" {
		t, err := time.Parse("2006-01-02", in.BirthDate)
		if err != nil {
			return nil, domain.ErrInvalidProfile
		}
		birthDate = &t
	}

	customer.FirstName = in.FirstName
	customer.LastName = in.LastName
	customer.BirthDate = birthDate
	customer.Address = in.Address
	customer.Phone = in.Phone
	customer.Citizenship = in.Citizenship

	// NEW → PROFILE_FILLED
	if customer.Status == roles.StatusNew {
		customer.Status = roles.StatusProfileFilled
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return nil, err
	}

	// Публикуем событие
	_ = uc.events.PublishProfileFilled(customer.ID, customer.UserID)

	return customer, nil
}
