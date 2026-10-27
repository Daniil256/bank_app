package domain

import "errors"

var (
	ErrNotFound           = errors.New("customer not found")
	ErrAccessDenied       = errors.New("access denied")
	ErrCannotEditOnKyc    = errors.New("cannot edit profile after KYC started")
	ErrInvalidProfile     = errors.New("invalid profile data")
	ErrCustomerIDRequired = errors.New("customer_id required")
)
