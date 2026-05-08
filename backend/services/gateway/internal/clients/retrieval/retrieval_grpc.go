package retrieval

import (
	"context"
	"fmt"

	retrievalv1 "github.com/keshvan/rag-as-service/backend/pkg/common/gen/retrieval/v1"
	commonClient "github.com/keshvan/rag-as-service/backend/pkg/common/grpc/client"
	"google.golang.org/grpc"
)

type Client struct {
	conn *grpc.ClientConn
	api  retrievalv1.RetrievalServiceClient
}

func New(host string, port int) (*Client, error) {
	conn, err := commonClient.NewGRPCClient(host, port)
	if err != nil {
		return nil, fmt.Errorf("failed to init auth grpc client: %w", err)
	}

	return &Client{
		conn: conn,
		api:  retrievalv1.NewRetrievalServiceClient(conn),
	}, nil
}

func (c *Client) Search(ctx context.Context, req *retrievalv1.SearchRequest) (*retrievalv1.SearchResponse, error) {
	return c.api.Search(ctx, req)
}

func (c *Client) Close() error {
	return c.conn.Close()
}
