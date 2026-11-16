package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"kyc-service/internal/domain"

	"github.com/lib/pq"
)

type ApplicationRepository struct {
	db *sql.DB
}

func NewApplicationRepository(db *sql.DB) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

// Create — вставляет заявку, ловит unique_violation на активной заявке
func (r *ApplicationRepository) Create(ctx context.Context, a *domain.Application) error {
	query := `
		INSERT INTO applications (user_id, status)
		VALUES ($1, $2)
	`

	err := r.db.QueryRowContext(ctx, query, a.UserID, a.Status).Scan()
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return domain.ErrActiveExists
		}
		return fmt.Errorf("create application: %w", err)
	}
	return nil
}

// GetByID — возвращает заявку по id
func (r *ApplicationRepository) GetByID(ctx context.Context, id string) (*domain.Application, error) {
	query := `
		SELECT id, user_id, status, claimed_by, rejection_reason, created_at, updated_at
		FROM applications WHERE id = $1
	`
	var a domain.Application
	var claimedBy, rejectionReason *string
	var updatedAt *time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.UserID, &a.Status,
		&claimedBy, &rejectionReason, &a.CreatedAt, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get application: %w", err)
	}
	if claimedBy != nil {
		a.ClaimedBy = claimedBy
	}
	if rejectionReason != nil {
		a.RejectionReason = rejectionReason
	}
	if updatedAt != nil {
		a.UpdatedAt = *updatedAt
	}
	return &a, nil
}

func (r *ApplicationRepository) GetByUserID(ctx context.Context, userId string) (*domain.Application, error) {
	query := `
		SELECT id, user_id, status, claimed_by, rejection_reason, created_at, updated_at
		FROM applications WHERE user_id = $1
	`
	var a domain.Application
	var claimedBy, rejectionReason *string
	var updatedAt *time.Time
	err := r.db.QueryRowContext(ctx, query, userId).Scan(
		&a.ID, &a.UserID, &a.Status,
		&claimedBy, &rejectionReason, &a.CreatedAt, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get application by user_id: %w", err)
	}
	if claimedBy != nil {
		a.ClaimedBy = claimedBy
	}
	if rejectionReason != nil {
		a.RejectionReason = rejectionReason
	}
	if updatedAt != nil {
		a.UpdatedAt = *updatedAt
	}
	return &a, nil
}

func (r *ApplicationRepository) List(ctx context.Context, status string, limit int, cursor string) ([]domain.Application, string, error) {
	query := `
		SELECT id, user_id, status, claimed_by, rejection_reason, created_at, updated_at
		FROM applications
		WHERE 1=1
	`
	args := []interface{}{}
	i := 1

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, status)
		i++
	}
	if cursor != "" {
		query += fmt.Sprintf(" AND id > $%d", i)
		args = append(args, cursor)
		i++
	}
	query += fmt.Sprintf(" ORDER BY id ASC LIMIT $%d", i)
	args = append(args, limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("list applications: %w", err)
	}
	defer rows.Close()

	var out []domain.Application
	for rows.Next() {
		var a domain.Application
		var claimedBy, rejectionReason *string
		var updatedAt *time.Time
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.Status,
			&claimedBy, &rejectionReason, &a.CreatedAt, &updatedAt,
		); err != nil {
			return nil, "", fmt.Errorf("scan application: %w", err)
		}
		if claimedBy != nil {
			a.ClaimedBy = claimedBy
		}
		if rejectionReason != nil {
			a.RejectionReason = rejectionReason
		}
		if updatedAt != nil {
			a.UpdatedAt = *updatedAt
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	nextToken := ""
	if len(out) > limit {
		nextToken = out[limit-1].ID
		out = out[:limit]
	}

	if out == nil {
		out = []domain.Application{}
	}
	return out, nextToken, nil
}

// Claim — оптимистическая блокировка: только если claimed_by IS NULL
func (r *ApplicationRepository) Claim(ctx context.Context, id, officerID string) (bool, error) {
	query := `
		UPDATE applications
		SET claimed_by = $1, claimed_at = NOW(), status = 'IN_REVIEW', updated_at = NOW()
		WHERE id = $2 AND claimed_by IS NULL AND status = 'IN_REVIEW'
	`
	res, err := r.db.ExecContext(ctx, query, officerID, id)
	if err != nil {
		return false, fmt.Errorf("claim application: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

// Approve — только если claimed_by = officerID и status = IN_REVIEW
func (r *ApplicationRepository) Approve(ctx context.Context, id, officerID string) (bool, error) {
	query := `
		UPDATE applications
		SET status = 'APPROVED', updated_at = NOW()
		WHERE id = $1 AND claimed_by = $2 AND status = 'IN_REVIEW'
	`
	res, err := r.db.ExecContext(ctx, query, id, officerID)
	if err != nil {
		return false, fmt.Errorf("approve application: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

// ListPending — заявки в статусе SUBMITTED
func (r *ApplicationRepository) ListPending(ctx context.Context, limit int) ([]domain.Application, error) {
	query := `
		SELECT id, user_id, status, claimed_by, rejection_reason, created_at, updated_at
		FROM applications
		WHERE status = 'SUBMITTED'
		ORDER BY created_at ASC
		LIMIT $1
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("list pending: %w", err)
	}
	defer rows.Close()

	var out []domain.Application
	for rows.Next() {
		var a domain.Application
		var claimedBy, rejectionReason *string
		var updatedAt *time.Time
		if err := rows.Scan(&a.ID, &a.UserID, &a.Status, &claimedBy, &rejectionReason, &a.CreatedAt, &updatedAt); err != nil {
			return nil, err
		}
		if claimedBy != nil {
			a.ClaimedBy = claimedBy
		}
		if rejectionReason != nil {
			a.RejectionReason = rejectionReason
		}
		if updatedAt != nil {
			a.UpdatedAt = *updatedAt
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// UpdateStatus — меняет статус
func (r *ApplicationRepository) UpdateStatus(ctx context.Context, id string, status domain.Status) error {
	query := `UPDATE applications SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}

// Reject — только если claimed_by = officerID и status = IN_REVIEW
func (r *ApplicationRepository) Reject(ctx context.Context, id, officerID, reason string) (bool, error) {
	query := `
		UPDATE applications
		SET status = 'REJECTED', rejection_reason = $3, updated_at = NOW()
		WHERE id = $1 AND claimed_by = $2 AND status = 'IN_REVIEW'
	`
	res, err := r.db.ExecContext(ctx, query, id, officerID, reason)
	if err != nil {
		return false, fmt.Errorf("reject application: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}
