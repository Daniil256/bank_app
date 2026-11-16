package customer

import (
	"context"
	"kyc-service/internal/domain"
)

type ProfileAdapter struct {
	client *Client
}

func (a *ProfileAdapter) GetCustomerProfile(ctx context.Context, customerID string) (*domain.CustomerProfile, error) {
	return a.client.GetProfile(ctx, customerID)
}
