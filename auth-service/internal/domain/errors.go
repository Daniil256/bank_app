package domain

import "errors"

var (
	ErrEmailExists         = errors.New("email already exists")
	ErrInvalidEmail        = errors.New("the email address is incorrect")
	ErrWeakPassword        = errors.New("password too weak")
	ErrInvalidCredentials  = errors.New("invalid username or password")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidLogout       = errors.New("invalid logout")
)
