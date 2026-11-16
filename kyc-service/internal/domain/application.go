package domain

import (
	"context"
	"time"
)

type Application struct {
	ID              string
	UserID          string
	Status          Status
	ClaimedBy       *string
	RejectionReason *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
type CustomerProfile struct {
	UserID    string
	FirstName string
	LastName  string
	BirthDate time.Time
}
type CustomerClient interface {
	GetProfile(ctx context.Context, userID string) (*CustomerProfile, error)
	GetStatus(ctx context.Context, customerID string) (string, error)
}
type ApplicationRepository interface {
	Create(ctx context.Context, a *Application) error
	GetByID(ctx context.Context, id string) (*Application, error)
	UpdateStatus(ctx context.Context, id string, status Status) error
	ListPending(ctx context.Context, limit int) ([]Application, error)
	List(ctx context.Context, status string, limit int, cursor string) ([]Application, string, error)
	Claim(ctx context.Context, id, officerID string) (bool, error)
	Approve(ctx context.Context, id, officerID string) (bool, error)
	Reject(ctx context.Context, id, officerID, reason string) (bool, error)
}
type ApplicationRepo interface {
	GetByID(ctx context.Context, id string) (*Application, error)
	UpdateStatus(ctx context.Context, id string, status Status) error
	ListPending(ctx context.Context, limit int) ([]Application, error)
	Reject(ctx context.Context, id, officerID, reason string) (bool, error)
}
