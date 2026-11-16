package usecase

import (
	"context"

	"kyc-service/internal/domain"
)

type ApproveRepository interface {
	ApplicationRepository
	Approve(ctx context.Context, id, officerID string) (bool, error)
}

type EventPublisher interface {
	PublishApplicationSubmitted(applicationID, userID string) error
	PublishApplicationApproved(applicationID, userID string) error
	PublishApplicationRejected(applicationID, userID, reason string) error
}

type ApproveUseCase struct {
	repo      ApproveRepository
	publisher EventPublisher
}

func NewApproveUseCase(repo ApproveRepository, publisher EventPublisher) *ApproveUseCase {
	return &ApproveUseCase{repo: repo, publisher: publisher}
}

func (uc *ApproveUseCase) Execute(ctx context.Context, applicationID, officerID string) (*domain.Application, error) {
	// 1. Проверяем, что заявка существует и в статусе IN_REVIEW
	app, err := uc.repo.GetByID(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, domain.ErrNotFound
	}
	if app.Status != domain.StatusInReview {
		return nil, domain.ErrNotInReview
	}

	// 2. Проверяем, что заявку взял именно этот офицер
	if app.ClaimedBy == nil || *app.ClaimedBy != officerID {
		return nil, domain.ErrNotClaimedByYou
	}

	// 3. Оптимистическая блокировка: UPDATE ... WHERE claimed_by = officerID
	ok, err := uc.repo.Approve(ctx, applicationID, officerID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrWrongStatus
	}

	// 4. Публикуем событие в Kafka
	_ = uc.publisher.PublishApplicationApproved(applicationID, app.UserID)

	// 5. Возвращаем обновлённую заявку
	return uc.repo.GetByID(ctx, applicationID)
}
