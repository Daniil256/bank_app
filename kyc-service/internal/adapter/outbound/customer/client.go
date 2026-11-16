package customer

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"kyc-service/internal/domain"
	customerv1 "kyc-service/proto/external/customer/v1"
	"time"
)

type Client struct {
	conn   *grpc.ClientConn
	client customerv1.CustomerServiceClient
}

func (c *Client) GetStatus(ctx context.Context, UserId string) (string, error) {
	request := customerv1.GetCustomerStatusRequest{
		UserId: UserId,
	}
	status, err := c.client.SystemGetCustomerStatus(ctx, &request)
	if err != nil {
		return "", err
	}
	return status.Status, nil
}

func NewClient(endpoint string) (*Client, error) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial customer: %w", err)
	}
	return &Client{
		conn:   conn,
		client: customerv1.NewCustomerServiceClient(conn),
	}, nil
}

func (c *Client) Close() error { return c.conn.Close() }

func (c *Client) GetProfile(ctx context.Context, userID string) (*domain.CustomerProfile, error) {
	resp, err := c.client.SystemGetCustomer(ctx, &customerv1.GetCustomerRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("system get customer status: %w", err)
	}

	birthDate, _ := time.Parse("2006-01-02", resp.BirthDate)
	return &domain.CustomerProfile{
		UserID:    resp.UserId,
		FirstName: resp.FirstName,
		LastName:  resp.LastName,
		BirthDate: birthDate,
	}, nil
}
