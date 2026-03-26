package auth

import (
	"fmt"

	commonClient "github.com/keshvan/rag-as-service/backend/pkg/common/grpc/client"
	"google.golang.org/grpc"
)

type Client struct {
	conn *grpc.ClientConn
	//api pb.AuthServiceClient protobuf
}

func New(host string, port int) (*Client, error) {
	conn, err := commonClient.NewGRPCClient(host, port)
	if err != nil {
		return nil, fmt.Errorf("failed to init auth grpc client: %w", err)
	}

	return &Client{
		conn: conn,
		//api: pb.NewAuthServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

//TODO...
