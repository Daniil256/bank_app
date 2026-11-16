package usecase

import (
	"context"
	"fmt"
	"kyc-service/internal/domain"
)

type ApplicationRepository interface {
	Create(ctx context.Context, a *domain.Application) error
	GetByID(ctx context.Context, id string) (*domain.Application, error)
	GetByUserID(ctx context.Context, userId string) (*domain.Application, error)
}
type DocumentChecker interface {
	HasRequired(ctx context.Context, customerID string) (bool, error)
}

type SubmitUseCase struct {
	repo      ApplicationRepository
	customer  domain.CustomerClient
	document  DocumentChecker
	publisher EventPublisher
}

func NewSubmitUseCase(
	repo ApplicationRepository,
	customer domain.CustomerClient,
	document DocumentChecker,
	publisher EventPublisher,
) *SubmitUseCase {
	return &SubmitUseCase{
		repo:      repo,
		customer:  customer,
		document:  document,
		publisher: publisher,
	}
}

func (uc *SubmitUseCase) Execute(ctx context.Context, userID string) (*domain.Application, error) {
	if userID == "" {
		return nil, domain.ErrCustomerIDRequired
	}

	// 1. Проверка профиля через customer-service (gRPC)
	status, err := uc.customer.GetStatus(ctx, userID)
	if err != nil {
		return nil, err
	}
	if status != "PROFILE_FILLED" {
		return nil, domain.ErrProfileNotFilled
	}

	// 2. Проверка документов через document-service (gRPC)
	fmt.Println("submit.document.HasRequired success")
	ok, err := uc.document.HasRequired(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrDocumentsMissing
	}

	// 3. Создание заявки
	app := &domain.Application{
		UserID: userID,
		Status: domain.StatusSubmitted,
	}
	if err := uc.repo.Create(ctx, app); err != nil {
		return nil, err
	}

	// 4. Публикация события в Kafka
	_ = uc.publisher.PublishApplicationSubmitted(app.ID, app.UserID)

	return app, nil
}
