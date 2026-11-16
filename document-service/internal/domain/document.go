package domain

import (
	"context"
	"time"
)

type DocumentStatus string

const (
	StatusUploaded  DocumentStatus = "UPLOADED"
	StatusConfirmed DocumentStatus = "CONFIRMED"
)

type Document struct {
	ID          string
	CustomerID  string
	DocType     DocType
	ObjectKey   string
	Status      DocumentStatus
	SizeBytes   int64
	MimeType    string
	SHA256      string
	CreatedAt   time.Time
	ConfirmedAt *time.Time
}

// DocumentRepository — репозиторий документов
type DocumentRepository interface {
	Create(ctx context.Context, d *Document) error
	GetByID(ctx context.Context, id string) (*Document, error)
	UpdateConfirmed(ctx context.Context, id string, size int64, mime, sha256 string) error
	ListByCustomer(ctx context.Context, customerID string, docType DocType) ([]Document, error) // ← новое
	HasConfirmedTypes(ctx context.Context, customerID string, types []DocType) (map[DocType]bool, error)
}

// StorageClient — интерфейс к MinIO
type StorageClient interface {
	PresignedPutURL(ctx context.Context, key string, ttl time.Duration) (string, error)
	PresignedGetURL(ctx context.Context, key string, ttl time.Duration) (string, error) // ← новое
	StatObject(ctx context.Context, key string) (size int64, err error)
	ReadHead(ctx context.Context, key string, n int) ([]byte, error)
	ReadFull(ctx context.Context, key string) ([]byte, error)
	RemoveObject(ctx context.Context, key string) error
}
