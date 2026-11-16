package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"kyc-service/internal/domain"
)

type CheckRepository struct {
	db *sql.DB
}

func NewCheckRepository(db *sql.DB) *CheckRepository {
	return &CheckRepository{db: db}
}

func (r *CheckRepository) Save(ctx context.Context, c *domain.KycCheck) error {
	query := `
		INSERT INTO kyc_checks (application_id, check_type, result, details)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query, c.ApplicationID, c.CheckType, c.Result, c.Details)
	if err != nil {
		return fmt.Errorf("save check: %w", err)
	}
	return nil
}
