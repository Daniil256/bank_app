package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"customer-service/internal/domain"

	"github.com/lib/pq"
)

type CustomerRepository struct {
	db *sql.DB
}

func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) GetByID(ctx context.Context, id string) (*domain.Customer, error) {
	query := `
		SELECT id, user_id, first_name, last_name, birth_date, address, phone, citizenship, status, created_at, updated_at
		FROM customers WHERE id = $1
	`
	var c domain.Customer
	var birthDate sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.UserID, &c.FirstName, &c.LastName, &birthDate,
		&c.Address, &c.Phone, &c.Citizenship, &c.Status, &c.CreatedAt, &c.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get customer: %w", err)
	}
	if birthDate.Valid {
		c.BirthDate = &birthDate.Time
	}
	return &c, nil
}

func (r *CustomerRepository) Create(ctx context.Context, c *domain.Customer) error {
	query := `
		INSERT INTO customers (user_id, status)
		VALUES ($1, 'NEW')
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, c.UserID).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return fmt.Errorf("customer for user %s already exists", c.UserID)
		}
		return fmt.Errorf("create customer: %w", err)
	}
	c.Status = "NEW"
	return nil
}

func (r *CustomerRepository) Update(ctx context.Context, c *domain.Customer) error {
	query := `
		UPDATE customers
		SET first_name = $1, last_name = $2, birth_date = $3, address = $4,
		    phone = $5, citizenship = $6, status = $7, updated_at = NOW()
		WHERE id = $8
	`
	_, err := r.db.ExecContext(ctx, query,
		c.FirstName, c.LastName, c.BirthDate, c.Address,
		c.Phone, c.Citizenship, c.Status, c.ID,
	)
	if err != nil {
		return fmt.Errorf("update customer: %w", err)
	}
	return nil
}

func (r *CustomerRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE customers SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}

func (r *CustomerRepository) List(ctx context.Context, f domain.ListFilter) (*domain.ListResult, error) {
	query := `
		SELECT id, user_id, first_name, last_name, birth_date, address, phone,
		       citizenship, status, created_at, updated_at
		FROM customers
	`
	args := []interface{}{}
	i := 1

	if f.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, f.Status)
		i++
	}
	if f.Search != "" {
		query += fmt.Sprintf(
			" AND (first_name ILIKE $%d OR last_name ILIKE $%d OR phone ILIKE $%d)",
			i, i+1, i+2,
		)
		like := "%" + f.Search + "%"
		args = append(args, like, like, like)
		i += 3
	}
	// keyset-пагинация по id (cursor = последний id предыдущей страницы)
	if f.Cursor != "" {
		query += fmt.Sprintf(" AND id > $%d", i)
		args = append(args, f.Cursor)
		i++
	}
	query += fmt.Sprintf(" ORDER BY id ASC LIMIT $%d", i)
	args = append(args, f.Limit+1) // +1 чтобы понять, есть ли следующая страница

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list customers: %w", err)
	}
	defer rows.Close()

	var out []domain.Customer
	for rows.Next() {
		var c domain.Customer
		var birthDate sql.NullTime
		if err := rows.Scan(
			&c.ID, &c.UserID, &c.FirstName, &c.LastName, &birthDate,
			&c.Address, &c.Phone, &c.Citizenship, &c.Status, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan customer: %w", err)
		}
		if birthDate.Valid {
			c.BirthDate = &birthDate.Time
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	res := &domain.ListResult{}
	if len(out) > f.Limit {
		res.NextToken = out[f.Limit-1].ID
		out = out[:f.Limit]
	}
	res.Customers = out
	return res, nil
}
