package domain

import "context"

type DocumentChecker interface {
	HasRequired(ctx context.Context, customerID string) (bool, error)
}
