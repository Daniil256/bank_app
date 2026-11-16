package worker

import (
	"context"
	"log"
	"time"

	"kyc-service/internal/domain"
)

type CheckRepository interface {
	Save(ctx context.Context, c *domain.KycCheck) error
}

type SanctionsChecker interface {
	Check(ctx context.Context, firstName, lastName, birthDate string) (bool, string, error)
}

type AutoChecksWorker struct {
	repo      domain.ApplicationRepo
	checks    CheckRepository
	sanctions SanctionsChecker
	customer  domain.CustomerClient
	interval  time.Duration
}

func NewAutoChecksWorker(
	repo domain.ApplicationRepo,
	checks CheckRepository,
	sanctions SanctionsChecker,
	customer domain.CustomerClient, // ← 4-й аргумент
) *AutoChecksWorker {
	return &AutoChecksWorker{
		repo:      repo,
		checks:    checks,
		sanctions: sanctions,
		customer:  customer,
		interval:  5 * time.Second,
	}
}

// Run — бесконечный цикл, пока ctx не отменён
func (w *AutoChecksWorker) Run(ctx context.Context) {
	log.Println("[auto_checks] worker started")
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[auto_checks] worker stopped")
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

// processBatch — берёт до 10 заявок SUBMITTED и обрабатывает
func (w *AutoChecksWorker) processBatch(ctx context.Context) {
	apps, err := w.repo.ListPending(ctx, 10)
	if err != nil {
		log.Printf("[auto_checks] list pending err: %v", err)
		return
	}
	if len(apps) == 0 {
		log.Println("[auto_checks] applications not found")
		return
	}
	log.Printf("[auto_checks] processing %d applications", len(apps))

	for _, app := range apps {
		if err := w.processOne(ctx, app); err != nil {
			log.Printf("[auto_checks] app=%s err: %v", app.ID, err)
		}
	}
}

// processOne — обработка одной заявки
func (w *AutoChecksWorker) processOne(ctx context.Context, app domain.Application) error {
	log.Printf("[auto_checks] app=%s start", app.ID)

	// 1. Статус → AUTO_CHECKS
	if err := w.repo.UpdateStatus(ctx, app.ID, domain.StatusAutoChecks); err != nil {
		return err
	}

	// 2. Получаем профиль клиента (ФИО, дата рождения) из customer-service
	profile, err := w.customer.GetProfile(ctx, app.UserID)

	if err != nil {
		log.Printf("[auto_checks] app=%s customer profile err: %v", app.ID, err)
		// Если профиль недоступен — не можем проверить, оставляем в AUTO_CHECKS
		return err
	}

	// 3. Проверка санкций
	sanctionsOK, reason, err := w.sanctions.Check(ctx, profile.FirstName, profile.LastName, profile.BirthDate.Format("2006-01-02"))
	if err != nil {
		log.Printf("[auto_checks] app=%s sanctions err: %v", app.ID, err)
		return err
	}
	result := domain.CheckPass
	if !sanctionsOK {
		result = domain.CheckFail
	}
	_ = w.checks.Save(ctx, &domain.KycCheck{
		ApplicationID: app.ID,
		CheckType:     "sanctions",
		Result:        result,
		Details:       reason,
	})
	if !sanctionsOK {
		log.Printf("[auto_checks] app=%s AUTO_REJECTED (sanctions: %s)", app.ID, reason)
		return w.repo.UpdateStatus(ctx, app.ID, domain.StatusAutoRejected)
	}

	// 4. Проверка возраста
	ageOK := CheckAge(profile.BirthDate)
	result = domain.CheckPass
	details := ""
	if !ageOK {
		result = domain.CheckFail
		details = "age < 18"
	}
	_ = w.checks.Save(ctx, &domain.KycCheck{
		ApplicationID: app.ID,
		CheckType:     "age",
		Result:        result,
		Details:       details,
	})
	if !ageOK {
		log.Printf("[auto_checks] app=%s AUTO_REJECTED (age < 18)", app.ID)
		return w.repo.UpdateStatus(ctx, app.ID, domain.StatusAutoRejected)
	}

	// 5. Все PASS → IN_REVIEW
	log.Printf("[auto_checks] app=%s all checks PASS → IN_REVIEW", app.ID)
	return w.repo.UpdateStatus(ctx, app.ID, domain.StatusInReview)
}
