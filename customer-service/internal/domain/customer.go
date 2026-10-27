package domain

import "time"

type Customer struct {
	ID          string
	UserID      string
	FirstName   string
	LastName    string
	BirthDate   *time.Time
	Address     string
	Phone       string
	Citizenship string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
