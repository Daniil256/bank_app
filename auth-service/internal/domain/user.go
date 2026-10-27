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
const RoleAdmin = "ADMIN"
const RoleOfficer = "OFFICER"

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}
