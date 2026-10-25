package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"auth-service/internal/domain"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create — сохранить хеш refresh-токена
func (r *RefreshTokenRepository) Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.ExecContext(ctx, query, userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}
	return nil
}

// FindByHash — найти активный refresh по хешу
func (r *RefreshTokenRepository) FindByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	var t domain.RefreshToken
	var revokedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &revokedAt, &t.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find refresh token: %w", err)
	}
	if revokedAt.Valid {
		t.RevokedAt = &revokedAt.Time
	}
	return &t, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("error revoke refresh token: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error affected refresh token: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("refresh token not found")
	}
	return nil
}

// RevokeAllForUser — отозвать все токены пользователя (logout everywhere)
func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("revoke all refresh tokens: %w", err)
	}
	return nil
}

// DeleteExpired — очистка истёкших (для cron)
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked_at IS NOT NULL`,
	)
	return err
}
