package auth

import (
	"fmt"

	commonClient "github.com/keshvan/rag-as-service/backend/pkg/common/grpc/client"
)

type Client struct {
	//protobuf
}

func New(host string, port int) (*Client, error) {
	_, err := commonClient.NewGRPCClient(host, port) // conn, err := ...
	if err != nil {
		return nil, fmt.Errorf("failed to init auth grpc client: %w", err)
	}

	return &Client{
		//api: pb.NewAuthServiceClient(conn),
	}, nil
}

//TODO...
