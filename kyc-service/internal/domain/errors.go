package domain

import "errors"

var (
	ErrNotFound           = errors.New("application not found")
	ErrActiveExists       = errors.New("active application already exists")
	ErrProfileNotFilled   = errors.New("profile must be filled")
	ErrDocumentsMissing   = errors.New("required documents missing")
	ErrAccessDenied       = errors.New("access denied")
	ErrCustomerIDRequired = errors.New("user_id required")
)
var (
	ErrNotInReview     = errors.New("application in not review")
	ErrAlreadyClaimed  = errors.New("application already claimed")
	ErrNotClaimedByYou = errors.New("application claimed by another officer")
	ErrWrongStatus     = errors.New("invalid status transition")
)
var (
	ErrReasonRequired = errors.New("rejection reason required")
)
