package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"auth-service/internal/domain"

	"github.com/lib/pq"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(ctx, query, u.Email, u.PasswordHash, u.Role).
		Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return domain.ErrEmailExists
		}
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email,
	).Scan(&exists)
	return exists, err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, role, created_at
		FROM users
		WHERE email = $1
	`
	var u domain.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	return &u, nil
}
func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	fmt.Println("[repo] FindByID userId ", id)

	query := `
		SELECT id, email, password_hash, role, customer_id, created_at
		FROM users
		WHERE id = $1
	`
	var u domain.User

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CustomerID, &u.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		fmt.Println("[repo] FindByID user not found. userId ", id)
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &u, nil
}
