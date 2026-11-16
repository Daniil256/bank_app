package usecase

import (
	"context"

	"document-service/internal/domain"
)

type ListDocumentsUseCase struct {
	repo domain.DocumentRepository
}

func NewListDocumentsUseCase(repo domain.DocumentRepository) *ListDocumentsUseCase {
	return &ListDocumentsUseCase{repo: repo}
}

type ListDocumentsInput struct {
	RequesterRole       string
	RequesterCustomerID string
	QueryCustomerID     string
	DocType             string
}

func (uc *ListDocumentsUseCase) Execute(ctx context.Context, in ListDocumentsInput) ([]domain.Document, error) {
	var targetCustomerID string

	switch in.RequesterRole {
	case "CLIENT":
		// CLIENT — только свой (игнорируем query)
		targetCustomerID = in.RequesterCustomerID
	case "OFFICER", "ADMIN":
		// OFFICER/ADMIN — только по customer_id из query
		if in.QueryCustomerID == "" {
			return nil, domain.ErrCustomerIDRequired
		}
		targetCustomerID = in.QueryCustomerID
	default:
		return nil, domain.ErrAccessDenied
	}

	if targetCustomerID == "" {
		return nil, domain.ErrCustomerIDRequired
	}

	docType := domain.DocType(in.DocType)
	if docType != "" && !docType.IsValid() {
		return nil, domain.ErrInvalidDocType
	}

	return uc.repo.ListByCustomer(ctx, targetCustomerID, docType)
}
