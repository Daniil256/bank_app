package grpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"document-service/internal/app/usecase"
	"document-service/internal/domain"
	documentv1 "document-service/proto/document/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type DocumentServer struct {
	documentv1.UnimplementedDocumentServiceServer
	initUpload     *usecase.InitUploadUseCase
	confirmUpload  *usecase.ConfirmUploadUseCase
	getDownloadURL *usecase.GetDownloadURLUseCase
	listDocuments  *usecase.ListDocumentsUseCase
	hasRequired    *usecase.HasRequiredDocumentsUseCase // ← новое
}

func NewDocumentServer(
	initUpload *usecase.InitUploadUseCase,
	confirmUpload *usecase.ConfirmUploadUseCase,
	getDownloadURL *usecase.GetDownloadURLUseCase,
	listDocuments *usecase.ListDocumentsUseCase,
	hasRequired *usecase.HasRequiredDocumentsUseCase, // ← новое
) *DocumentServer {
	return &DocumentServer{
		initUpload:     initUpload,
		confirmUpload:  confirmUpload,
		getDownloadURL: getDownloadURL,
		listDocuments:  listDocuments,
		hasRequired:    hasRequired,
	}
}

func (s *DocumentServer) InitUpload(ctx context.Context, req *documentv1.InitUploadRequest) (*documentv1.InitUploadResponse, error) {
	fmt.Println("[document.InitUpload] req", req)

	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")
	userID := getMeta(md, "x-user-id")

	if role != "CLIENT" {
		fmt.Println("[document.InitUpload] err: only CLIENT")
		return nil, status.Error(codes.PermissionDenied, "only CLIENT can upload")
	}
	if userID == "" {
		fmt.Println("[document.InitUpload] err: user-id missing")
		return nil, status.Error(codes.FailedPrecondition, "customer_id required")
	}

	out, err := s.initUpload.Execute(ctx, usecase.InitUploadInput{
		UserID:  userID,
		DocType: req.DocType,
	})
	if err != nil {
		fmt.Println("[document.InitUpload] Execute err", err)
		return nil, mapError(err)
	}

	fmt.Println("[document.InitUpload] success document_id=", out.DocumentID)
	return &documentv1.InitUploadResponse{
		DocumentId: out.DocumentID,
		UploadUrl:  out.UploadURL,
		ObjectKey:  out.ObjectKey,
	}, nil
}

func (s *DocumentServer) ConfirmUpload(ctx context.Context, req *documentv1.ConfirmUploadRequest) (*documentv1.Document, error) {
	fmt.Println("[document.ConfirmUpload] req", req)

	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")
	userID := getMeta(md, "x-user-id")

	if userID == "" {
		fmt.Println("[document.ConfirmUpload] err: user-id missing")
		return nil, status.Error(codes.FailedPrecondition, "user-id required")
	}

	if role != "CLIENT" {
		fmt.Println("[document.ConfirmUpload] err: only CLIENT")
		return nil, status.Error(codes.PermissionDenied, "only CLIENT can confirm")
	}

	doc, err := s.confirmUpload.Execute(ctx, req.DocumentId, userID)
	if err != nil {
		fmt.Println("[document.ConfirmUpload] Execute err", err)
		return nil, mapError(err)
	}

	fmt.Println("[document.ConfirmUpload] success document_id=", doc.ID)
	return toProto(doc), nil
}

func getMeta(md metadata.MD, key string) string {
	v := md.Get(key)
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

func toProto(d *domain.Document) *documentv1.Document {
	p := &documentv1.Document{
		Id:         d.ID,
		CustomerId: d.CustomerID,
		DocType:    string(d.DocType),
		Status:     string(d.Status),
		SizeBytes:  d.SizeBytes,
		MimeType:   d.MimeType,
		Sha256:     d.SHA256,
		CreatedAt:  d.CreatedAt.Format(time.RFC3339),
	}
	if d.ConfirmedAt != nil {
		p.ConfirmedAt = d.ConfirmedAt.Format(time.RFC3339)
	}
	return p
}

func (s *DocumentServer) GetDownloadUrl(ctx context.Context, req *documentv1.GetDownloadUrlRequest) (*documentv1.GetDownloadUrlResponse, error) {
	fmt.Println("[document.GetDownloadUrl] req", req)

	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")
	customerID := getMeta(md, "x-user-id")

	out, err := s.getDownloadURL.Execute(ctx, usecase.GetDownloadURLInput{
		DocumentID:          req.DocumentId,
		RequesterRole:       role,
		RequesterCustomerID: customerID,
	})
	if err != nil {
		fmt.Println("[document.GetDownloadUrl] Execute err", err)
		return nil, mapError(err)
	}

	fmt.Println("[document.GetDownloadUrl] success document_id=", req.DocumentId)
	return &documentv1.GetDownloadUrlResponse{
		DownloadUrl: out.DownloadURL,
		ExpiresAt:   out.ExpiresAt.Format(time.RFC3339),
	}, nil
}

func (s *DocumentServer) ListDocuments(ctx context.Context, req *documentv1.ListDocumentsRequest) (*documentv1.ListDocumentsResponse, error) {
	fmt.Println("[document.ListDocuments] req", req)

	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")
	customerID := getMeta(md, "x-user-id")

	docs, err := s.listDocuments.Execute(ctx, usecase.ListDocumentsInput{
		RequesterRole:       role,
		RequesterCustomerID: customerID,
		QueryCustomerID:     req.CustomerId,
		DocType:             req.DocType,
	})
	if err != nil {
		fmt.Println("[document.ListDocuments] Execute err", err)
		return nil, mapError(err)
	}

	out := make([]*documentv1.Document, 0, len(docs))
	for i := range docs {
		out = append(out, toProto(&docs[i]))
	}

	fmt.Println("[document.ListDocuments] success count=", len(out))
	return &documentv1.ListDocumentsResponse{Documents: out}, nil
}

func (s *DocumentServer) HasRequiredDocuments(ctx context.Context, req *documentv1.HasRequiredDocumentsRequest) (*documentv1.HasRequiredDocumentsResponse, error) {
	fmt.Println("[document.HasRequiredDocuments] req", req)

	out, err := s.hasRequired.Execute(ctx, usecase.HasRequiredDocumentsInput{
		CustomerID: req.CustomerId,
	})
	if err != nil {
		fmt.Println("[document.HasRequiredDocuments] Execute err", err)
		return nil, mapError(err)
	}

	fmt.Println("[document.HasRequiredDocuments] success customer_id=", req.CustomerId,
		"has_passport=", out.HasPassport, "has_selfie=", out.HasSelfie)

	return &documentv1.HasRequiredDocumentsResponse{
		HasPassport: out.HasPassport,
		HasSelfie:   out.HasSelfie,
	}, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, "document not found")
	case errors.Is(err, domain.ErrInvalidDocType):
		return status.Error(codes.InvalidArgument, "invalid document type")
	case errors.Is(err, domain.ErrNotUploaded):
		return status.Error(codes.FailedPrecondition, "file not uploaded to storage")
	case errors.Is(err, domain.ErrAlreadyConfirmed):
		return status.Error(codes.FailedPrecondition, "document already confirmed")
	case errors.Is(err, domain.ErrFileTooLarge):
		return status.Error(codes.InvalidArgument, "file too large (max 10 MB)")
	case errors.Is(err, domain.ErrInvalidMimeType):
		return status.Error(codes.InvalidArgument, "invalid mime type (jpeg/png/pdf only)")
	case errors.Is(err, domain.ErrCustomerIDRequired):
		return status.Error(codes.FailedPrecondition, "customer_id required")
	case errors.Is(err, domain.ErrAccessDenied):
		return status.Error(codes.PermissionDenied, "access denied")
	case errors.Is(err, domain.ErrNotConfirmed):
		return status.Error(codes.FailedPrecondition, "document not confirmed")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
