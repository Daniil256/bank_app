package usecase

import (
	"context"
	"strings"

	"kyc-service/internal/domain"
)

type RejectRepository interface {
	domain.ApplicationRepository
	Reject(ctx context.Context, id, officerID, reason string) (bool, error)
}

type RejectUseCase struct {
	repo      RejectRepository
	publisher EventPublisher
}

func NewRejectUseCase(repo RejectRepository, publisher EventPublisher) *RejectUseCase {
	return &RejectUseCase{repo: repo, publisher: publisher}
}

func (uc *RejectUseCase) Execute(ctx context.Context, applicationID, officerID, reason string) (*domain.Application, error) {
	// 1. Причина обязательна
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, domain.ErrReasonRequired
	}

	// 2. Заявка существует?
	app, err := uc.repo.GetByID(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, domain.ErrNotFound
	}

	// 3. Статус IN_REVIEW?
	if app.Status != domain.StatusInReview {
		return nil, domain.ErrNotInReview
	}

	// 4. Заявку взял именно этот офицер?
	if app.ClaimedBy == nil || *app.ClaimedBy != officerID {
		return nil, domain.ErrNotClaimedByYou
	}

	// 5. Оптимистическая блокировка: UPDATE ... WHERE claimed_by = officerID
	ok, err := uc.repo.Reject(ctx, applicationID, officerID, reason)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrWrongStatus
	}

	// 6. Kafka: kyc.application_rejected
	_ = uc.publisher.PublishApplicationRejected(applicationID, app.UserID, reason)

	// 7. Возвращаем обновлённую заявку
	return uc.repo.GetByID(ctx, applicationID)
}
