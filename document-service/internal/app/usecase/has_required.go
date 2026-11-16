package usecase

import (
	"context"

	"document-service/internal/domain"
)

type HasRequiredDocumentsUseCase struct {
	repo domain.DocumentRepository
}

func NewHasRequiredDocumentsUseCase(repo domain.DocumentRepository) *HasRequiredDocumentsUseCase {
	return &HasRequiredDocumentsUseCase{repo: repo}
}

type HasRequiredDocumentsInput struct {
	CustomerID string
}

type HasRequiredDocumentsOutput struct {
	HasPassport bool
	HasSelfie   bool
}

func (uc *HasRequiredDocumentsUseCase) Execute(ctx context.Context, in HasRequiredDocumentsInput) (*HasRequiredDocumentsOutput, error) {
	if in.CustomerID == "" {
		return nil, domain.ErrCustomerIDRequired
	}

	requiredTypes := []domain.DocType{
		domain.DocTypePassport,
		domain.DocTypeSelfie,
	}

	confirmed, err := uc.repo.HasConfirmedTypes(ctx, in.CustomerID, requiredTypes)
	if err != nil {
		return nil, err
	}

	return &HasRequiredDocumentsOutput{
		HasPassport: confirmed[domain.DocTypePassport],
		HasSelfie:   confirmed[domain.DocTypeSelfie],
	}, nil
}
