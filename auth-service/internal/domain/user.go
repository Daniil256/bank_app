package domain

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
	CustomerID   string
	CreatedAt    time.Time
}

const RoleClient = "CLIENT"

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}
