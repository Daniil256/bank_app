package usecase

import (
	"context"
	"fmt"
	"time"

	"document-service/internal/domain"

	"github.com/google/uuid"
)

type InitUploadUseCase struct {
	repo    domain.DocumentRepository
	storage domain.StorageClient
	putTTL  time.Duration
}

func NewInitUploadUseCase(repo domain.DocumentRepository, storage domain.StorageClient) *InitUploadUseCase {
	return &InitUploadUseCase{
		repo:    repo,
		storage: storage,
		putTTL:  15 * time.Minute,
	}
}

type InitUploadInput struct {
	UserID  string
	DocType string
}

type InitUploadOutput struct {
	DocumentID string
	UploadURL  string
	ObjectKey  string
}

func (uc *InitUploadUseCase) Execute(ctx context.Context, in InitUploadInput) (*InitUploadOutput, error) {
	if in.UserID == "" {
		return nil, domain.ErrCustomerIDRequired
	}

	docType := domain.DocType(in.DocType)
	if !docType.IsValid() {
		return nil, domain.ErrInvalidDocType
	}

	objectKey := fmt.Sprintf("customers/%s/%s/%s", in.UserID, docType, uuid.NewString())

	doc := &domain.Document{
		CustomerID: in.UserID,
		DocType:    docType,
		ObjectKey:  objectKey,
		Status:     domain.StatusUploaded,
	}
	if err := uc.repo.Create(ctx, doc); err != nil {
		return nil, err
	}

	url, err := uc.storage.PresignedPutURL(ctx, objectKey, uc.putTTL)
	if err != nil {
		return nil, err
	}

	return &InitUploadOutput{
		DocumentID: doc.ID,
		UploadURL:  url,
		ObjectKey:  objectKey,
	}, nil
}
