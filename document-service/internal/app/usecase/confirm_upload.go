package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"document-service/internal/domain"
)

const maxFileSize = 10 * 1024 * 1024 // 10 MB

type ConfirmUploadUseCase struct {
	repo    domain.DocumentRepository
	storage domain.StorageClient
}

func NewConfirmUploadUseCase(repo domain.DocumentRepository, storage domain.StorageClient) *ConfirmUploadUseCase {
	return &ConfirmUploadUseCase{repo: repo, storage: storage}
}

func (uc *ConfirmUploadUseCase) Execute(ctx context.Context, documentID, userID string) (*domain.Document, error) {
	// 1. Загружаем документ
	doc, err := uc.repo.GetByID(ctx, documentID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, domain.ErrNotFound
	}

	// 2. Проверка владельца
	if doc.CustomerID != userID {
		return nil, domain.ErrAccessDenied
	}

	// 3. Проверка статуса
	if doc.Status != domain.StatusUploaded {
		return nil, domain.ErrAlreadyConfirmed
	}

	// 4. Проверка, что файл загружен в MinIO
	size, err := uc.storage.StatObject(ctx, doc.ObjectKey)
	if err != nil {
		return nil, domain.ErrNotUploaded
	}

	// 5. Проверка размера
	if size > maxFileSize {
		_ = uc.storage.RemoveObject(ctx, doc.ObjectKey)
		return nil, domain.ErrFileTooLarge
	}

	// 6. Проверка magic bytes
	head, err := uc.storage.ReadHead(ctx, doc.ObjectKey, 512)
	if err != nil {
		return nil, err
	}
	mime, ok := detectMime(head)
	if !ok {
		_ = uc.storage.RemoveObject(ctx, doc.ObjectKey)
		return nil, domain.ErrInvalidMimeType
	}

	// 7. Читаем полностью для sha256
	full, err := uc.storage.ReadFull(ctx, doc.ObjectKey)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(full)
	sha := hex.EncodeToString(sum[:])

	// 8. Обновляем БД
	if err := uc.repo.UpdateConfirmed(ctx, doc.ID, size, mime, sha); err != nil {
		return nil, err
	}

	// 9. Возвращаем обновлённый документ
	return uc.repo.GetByID(ctx, doc.ID)
}

// detectMime — проверка по magic bytes
func detectMime(b []byte) (string, bool) {
	if len(b) < 4 {
		return "", false
	}
	// JPEG: FF D8 FF
	if b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF {
		return "image/jpeg", true
	}
	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if len(b) >= 8 && b[0] == 0x89 && b[1] == 0x50 && b[2] == 0x4E && b[3] == 0x47 &&
		b[4] == 0x0D && b[5] == 0x0A && b[6] == 0x1A && b[7] == 0x0A {
		return "image/png", true
	}
	// PDF: 25 50 44 46
	if b[0] == 0x25 && b[1] == 0x50 && b[2] == 0x44 && b[3] == 0x46 {
		return "application/pdf", true
	}
	return "", false
}
