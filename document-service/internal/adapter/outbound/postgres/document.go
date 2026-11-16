package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"

	"document-service/internal/domain"
)

type DocumentRepository struct {
	db *sql.DB
}

func NewDocumentRepository(db *sql.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Create(ctx context.Context, d *domain.Document) error {
	query := `
		INSERT INTO documents (customer_id, doc_type, object_key, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(ctx, query, d.CustomerID, d.DocType, d.ObjectKey, d.Status).
		Scan(&d.ID, &d.CreatedAt)
	if err != nil {
		return fmt.Errorf("create document: %w", err)
	}
	return nil
}

func (r *DocumentRepository) GetByID(ctx context.Context, id string) (*domain.Document, error) {
	query := `
		SELECT id, customer_id, doc_type, object_key, status, size_bytes, mime_type, sha256, created_at, confirmed_at
		FROM documents WHERE id = $1
	`
	var d domain.Document
	var size sql.NullInt64
	var mime, sha sql.NullString
	var confirmedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID, &d.CustomerID, &d.DocType, &d.ObjectKey, &d.Status,
		&size, &mime, &sha, &d.CreatedAt, &confirmedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}
	if size.Valid {
		d.SizeBytes = size.Int64
	}
	if mime.Valid {
		d.MimeType = mime.String
	}
	if sha.Valid {
		d.SHA256 = sha.String
	}
	if confirmedAt.Valid {
		d.ConfirmedAt = &confirmedAt.Time
	}
	return &d, nil
}

func (r *DocumentRepository) UpdateConfirmed(ctx context.Context, id string, size int64, mime, sha256 string) error {
	query := `
		UPDATE documents
		SET status = 'CONFIRMED', size_bytes = $2, mime_type = $3, sha256 = $4, confirmed_at = NOW()
		WHERE id = $1 AND status = 'UPLOADED'
	`
	res, err := r.db.ExecContext(ctx, query, id, size, mime, sha256)
	if err != nil {
		return fmt.Errorf("update confirmed: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return domain.ErrAlreadyConfirmed
	}
	return nil
}

func (r *DocumentRepository) ListByCustomer(ctx context.Context, customerID string, docType domain.DocType) ([]domain.Document, error) {
	query := `
		SELECT id, customer_id, doc_type, object_key, status, size_bytes, mime_type, sha256, created_at, confirmed_at
		FROM documents
		WHERE customer_id = $1
	`
	args := []interface{}{customerID}

	if docType != "" {
		query += " AND doc_type = $2"
		args = append(args, docType)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	defer rows.Close()

	var out []domain.Document
	for rows.Next() {
		var d domain.Document
		var size sql.NullInt64
		var mime, sha sql.NullString
		var confirmedAt sql.NullTime

		if err := rows.Scan(
			&d.ID, &d.CustomerID, &d.DocType, &d.ObjectKey, &d.Status,
			&size, &mime, &sha, &d.CreatedAt, &confirmedAt,
		); err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		if size.Valid {
			d.SizeBytes = size.Int64
		}
		if mime.Valid {
			d.MimeType = mime.String
		}
		if sha.Valid {
			d.SHA256 = sha.String
		}
		if confirmedAt.Valid {
			d.ConfirmedAt = &confirmedAt.Time
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out == nil {
		out = []domain.Document{}
	}
	return out, nil
}
func (r *DocumentRepository) HasConfirmedTypes(ctx context.Context, customerID string, types []domain.DocType) (map[domain.DocType]bool, error) {
	// Инициализируем все false
	result := make(map[domain.DocType]bool, len(types))
	for _, t := range types {
		result[t] = false
	}

	if len(types) == 0 {
		return result, nil
	}

	// Преобразуем []DocType → []string для pq.Array
	typeStrs := make([]string, len(types))
	for i, t := range types {
		typeStrs[i] = string(t)
	}

	query := `
		SELECT DISTINCT doc_type
		FROM documents
		WHERE customer_id = $1
		  AND status = 'CONFIRMED'
		  AND doc_type = ANY($2)
	`
	rows, err := r.db.QueryContext(ctx, query, customerID, pq.Array(typeStrs))
	if err != nil {
		return nil, fmt.Errorf("has confirmed types: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, fmt.Errorf("scan doc_type: %w", err)
		}
		result[domain.DocType(t)] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

var _ domain.DocumentRepository = (*DocumentRepository)(nil)
