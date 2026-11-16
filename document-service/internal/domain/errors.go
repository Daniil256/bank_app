package domain

import "errors"

var (
	ErrInvalidDocType     = errors.New("invalid document type")
	ErrNotFound           = errors.New("document not found")
	ErrNotUploaded        = errors.New("document not uploaded")
	ErrAlreadyConfirmed   = errors.New("document already confirmed")
	ErrFileTooLarge       = errors.New("file too large (max 10 MB)")
	ErrInvalidMimeType    = errors.New("invalid mime type (jpeg/png/pdf only)")
	ErrCustomerIDRequired = errors.New("customer_id required")
	ErrNotConfirmed       = errors.New("document not confirmed")
	ErrAccessDenied       = errors.New("access denied")
)
