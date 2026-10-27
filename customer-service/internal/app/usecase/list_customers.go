package usecase

import (
	"context"

	"bank_app/shared/roles"
	"customer-service/internal/domain"
)

// CustomerListRepository — расширение репозитория для листинга
type CustomerListRepository interface {
	CustomerRepository
	List(ctx context.Context, f domain.ListFilter) (*domain.ListResult, error)
}

// ListCustomersUseCase — листинг клиентов (только OFFICER/ADMIN)
type ListCustomersUseCase struct {
	repo CustomerListRepository
}

func NewListCustomersUseCase(repo CustomerListRepository) *ListCustomersUseCase {
	return &ListCustomersUseCase{repo: repo}
}

// Execute — возвращает список клиентов с фильтрами и пагинацией
func (uc *ListCustomersUseCase) Execute(
	ctx context.Context,
	requesterRole string,
	filter domain.ListFilter,
) (*domain.ListResult, error) {
	// 1. Проверка роли — только OFFICER/ADMIN
	if requesterRole != roles.Officer && requesterRole != roles.Admin {
		return nil, domain.ErrAccessDenied
	}

	// 2. Нормализация фильтров
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	// 3. Валидация статуса (если задан)
	if filter.Status != "" && !isValidStatus(filter.Status) {
		return nil, domain.ErrInvalidProfile
	}

	// 4. Запрос в репозиторий
	return uc.repo.List(ctx, filter)
}

func isValidStatus(s string) bool {
	switch s {
	case roles.StatusNew,
		roles.StatusProfileFilled,
		roles.StatusOnKyc,
		roles.StatusActive,
		roles.StatusRejected,
		roles.StatusBlocked:
		return true
	}
	return false
}
