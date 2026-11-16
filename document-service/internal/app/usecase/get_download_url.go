package usecase

import (
	"context"
	"time"

	"document-service/internal/domain"
)

type GetDownloadURLUseCase struct {
	repo    domain.DocumentRepository
	storage domain.StorageClient
	getTTL  time.Duration
}

func NewGetDownloadURLUseCase(repo domain.DocumentRepository, storage domain.StorageClient) *GetDownloadURLUseCase {
	return &GetDownloadURLUseCase{
		repo:    repo,
		storage: storage,
		getTTL:  5 * time.Minute,
	}
}

type GetDownloadURLInput struct {
	DocumentID          string
	RequesterRole       string
	RequesterCustomerID string
}

type GetDownloadURLOutput struct {
	DownloadURL string
	ExpiresAt   time.Time
}

// Execute Проверять роль в самом начале. Добавить запрос GetByID для юзера для поиска doc_id и customer_id
func (uc *GetDownloadURLUseCase) Execute(ctx context.Context, in GetDownloadURLInput) (*GetDownloadURLOutput, error) {
	// 1. Загружаем документ
	doc, err := uc.repo.GetByID(ctx, in.DocumentID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, domain.ErrNotFound
	}

	// 2. Только CONFIRMED
	if doc.Status != domain.StatusConfirmed {
		return nil, domain.ErrNotConfirmed
	}

	// 3. Проверка доступа
	switch in.RequesterRole {
	case "CLIENT":
		// CLIENT — только свой
		if doc.CustomerID != in.RequesterCustomerID {
			return nil, domain.ErrAccessDenied
		}
	case "OFFICER", "ADMIN":
		// OFFICER/ADMIN — любой документ (проверка через kyc-заявку — на уровне gateway/kyc)
	default:
		return nil, domain.ErrAccessDenied
	}

	// 4. Presigned GET URL
	url, err := uc.storage.PresignedGetURL(ctx, doc.ObjectKey, uc.getTTL)
	if err != nil {
		return nil, err
	}

	return &GetDownloadURLOutput{
		DownloadURL: url,
		ExpiresAt:   time.Now().Add(uc.getTTL),
	}, nil
}
