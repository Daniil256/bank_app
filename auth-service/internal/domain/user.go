package domain

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}

const RoleClient = "CLIENT"
