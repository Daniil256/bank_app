package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type SanctionsChecker struct {
	db *sql.DB
}

func NewSanctionsChecker(db *sql.DB) *SanctionsChecker {
	return &SanctionsChecker{db: db}
}

// Check — ищет совпадение по ФИО + дате рождения
func (s *SanctionsChecker) Check(ctx context.Context, firstName, lastName, birthDate string) (bool, string, error) {
	query := `
		SELECT reason FROM sanctions_list
		WHERE first_name = $1 AND last_name = $2 AND birth_date = $3
		LIMIT 1
	`
	var reason string
	err := s.db.QueryRowContext(ctx, query, firstName, lastName, birthDate).Scan(&reason)
	if errors.Is(err, sql.ErrNoRows) {
		return true, "", nil // PASS — нет совпадений
	}
	if err != nil {
		return false, "", fmt.Errorf("sanctions check: %w", err)
	}
	return false, reason, nil // FAIL — совпадение
}
