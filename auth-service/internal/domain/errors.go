package domain

import "errors"

var (
	ErrEmailTaken         = errors.New("email already taken")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrWeakPassword       = errors.New("password too weak")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
