package auth

import (
	"context"
	"fmt"

	authv1 "github.com/keshvan/rag-as-service/backend/pkg/common/gen"
	commonClient "github.com/keshvan/rag-as-service/backend/pkg/common/grpc/client"
	"google.golang.org/grpc"
)

type Client struct {
	conn *grpc.ClientConn
	api  authv1.AuthServiceClient
}

func New(host string, port int) (*Client, error) {
	conn, err := commonClient.NewGRPCClient(host, port)
	if err != nil {
		return nil, fmt.Errorf("failed to init auth grpc client: %w", err)
	}

	return &Client{
		conn: conn,
		api:  authv1.NewAuthServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	return c.api.Register(ctx, req)
}

func (c *Client) VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest) (*authv1.VerifyEmailResponse, error) {
	return c.api.VerifyEmail(ctx, req)
}

func (c *Client) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	return c.api.Login(ctx, req)
}

func (c *Client) GetCurrentUserGUID(ctx context.Context, req *authv1.GetCurrentUserGUIDRequest) (*authv1.GetCurrentUserGUIDResponse, error) {
	return c.api.GetCurrentUserGUID(ctx, req)
}

func (c *Client) GetTokenPairByUserGUID(ctx context.Context, req *authv1.GetTokenPairByUserGUIDRequest) (*authv1.GetTokenPairByUserGUIDResponse, error) {
	return c.api.GetTokenPairByUserGUID(ctx, req)
}

func (c *Client) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	return c.api.Logout(ctx, req)
}

func (c *Client) RefreshTokens(ctx context.Context, req *authv1.RefreshTokensRequest) (*authv1.RefreshTokensResponse, error) {
	return c.api.RefreshTokens(ctx, req)
}
