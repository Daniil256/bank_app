package document

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	documentv1 "kyc-service/proto/external/document/v1"
)

type Client struct {
	conn   *grpc.ClientConn
	client documentv1.DocumentServiceClient
}

func NewClient(endpoint string) (*Client, error) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial document: %w", err)
	}
	return &Client{
		conn:   conn,
		client: documentv1.NewDocumentServiceClient(conn),
	}, nil
}

func (c *Client) Close() error { return c.conn.Close() }

// HasRequired — для submit use case
func (c *Client) HasRequired(ctx context.Context, customerID string) (bool, error) {
	resp, err := c.client.HasRequiredDocuments(ctx, &documentv1.HasRequiredDocumentsRequest{
		CustomerId: customerID,
	})
	if err != nil {
		return false, fmt.Errorf("has required documents: %w", err)
	}
	return resp.HasPassport && resp.HasSelfie, nil
}
