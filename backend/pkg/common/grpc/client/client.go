package client

import (
	"fmt"

	"github.com/keshvan/rag-as-service/backend/pkg/common/grpc/interceptors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGRPCClient(host string, port int) (*grpc.ClientConn, error) {
	addr := fmt.Sprintf("%s:%d", host, port)

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptors.ClientTenantInterceptor),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC service at %s: %w", addr, err)
	}

	return conn, nil
}
