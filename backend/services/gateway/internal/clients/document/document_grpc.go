package document

import (
	"context"
	"fmt"

	documentv1 "github.com/keshvan/rag-as-service/backend/pkg/common/gen/document/v1"
	commonClient "github.com/keshvan/rag-as-service/backend/pkg/common/grpc/client"
	"google.golang.org/grpc"
)

type Client struct {
	conn *grpc.ClientConn
	api  documentv1.DocumentServiceClient
}

func New(host string, port int) (*Client, error) {
	conn, err := commonClient.NewGRPCClient(host, port)
	if err != nil {
		return nil, fmt.Errorf("failed to init auth grpc client: %w", err)
	}

	return &Client{
		conn: conn,
		api:  documentv1.NewDocumentServiceClient(conn),
	}, nil
}

func (c *Client) InitUpload(ctx context.Context, req *documentv1.InitUploadRequest) (*documentv1.InitUploadResponse, error) {
	return c.api.InitUpload(ctx, req)
}

func (c *Client) CompleteUpload(ctx context.Context, req *documentv1.CompleteUploadRequest) (*documentv1.CompleteUploadResponse, error) {
	return c.api.CompleteUpload(ctx, req)
}

func (c *Client) Close() error {
	return c.conn.Close()
}

