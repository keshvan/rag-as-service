package grpcserver

import (
	"context"
	"errors"

	authv1 "github.com/keshvan/rag-as-service/backend/pkg/common/gen/auth/v1"
	"github.com/keshvan/rag-as-service/backend/services/auth/internal/authctx"
	"github.com/keshvan/rag-as-service/backend/services/auth/internal/lib/jwt"
	"github.com/keshvan/rag-as-service/backend/services/auth/internal/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	SendingEmailWithCode(ctx context.Context, email, password, orgName, orgURL string) error
	ConfirmVerificationCode(ctx context.Context, email, code string) (string, error)
	Login(ctx context.Context, email, password, userAgent, ip string) (jwt.TokenPair, string, error)
	GetCurrentUserGUID(ctx context.Context) (string, error)
	GetTokenPairByUserGUID(ctx context.Context, targetGUID, userAgent, ip string) (jwt.TokenPair, error)
	Logout(ctx context.Context) error
	RefreshTokens(ctx context.Context, refreshToken, userAgent, ip string) (jwt.TokenPair, error)
}

type AuthServer struct {
	authv1.UnimplementedAuthServiceServer
	auth AuthService
}

func NewAuthServer(authService AuthService) *AuthServer {
	return &AuthServer{auth: authService}
}

func (s *AuthServer) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" || req.GetOrganizationName() == "" || req.GetOrganizationUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, "missing required fields")
	}

	if err := s.auth.SendingEmailWithCode(ctx, req.GetEmail(), req.GetPassword(), req.GetOrganizationName(), req.GetOrganizationUrl()); err != nil {
		return nil, mapError(err)
	}

	return &authv1.RegisterResponse{
		Message: "If email is valid, a verification code has been sent",
	}, nil
}

func (s *AuthServer) VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest) (*authv1.VerifyEmailResponse, error) {
	if req.GetEmail() == "" || req.GetCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "missing required fields")
	}

	guid, err := s.auth.ConfirmVerificationCode(ctx, req.GetEmail(), req.GetCode())
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.VerifyEmailResponse{Guid: guid}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "missing required fields")
	}

	tokenPair, guid, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetUserAgent(), req.GetIp())
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.LoginResponse{
		Guid:         guid,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

func (s *AuthServer) GetCurrentUserGUID(ctx context.Context, req *authv1.GetCurrentUserGUIDRequest) (*authv1.GetCurrentUserGUIDResponse, error) {
	ctx = authctx.WithUserGUID(ctx, req.GetUserGuid())

	guid, err := s.auth.GetCurrentUserGUID(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.GetCurrentUserGUIDResponse{Guid: guid}, nil
}

func (s *AuthServer) GetTokenPairByUserGUID(ctx context.Context, req *authv1.GetTokenPairByUserGUIDRequest) (*authv1.GetTokenPairByUserGUIDResponse, error) {
	ctx = authctx.WithUserGUID(ctx, req.GetCurrentUserGuid())

	tokenPair, err := s.auth.GetTokenPairByUserGUID(
		ctx,
		req.GetTargetGuid(),
		req.GetUserAgent(),
		req.GetIp(),
	)
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.GetTokenPairByUserGUIDResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	ctx = authctx.WithUserGUID(ctx, req.GetUserGuid())

	if err := s.auth.Logout(ctx); err != nil {
		return nil, mapError(err)
	}

	return &authv1.LogoutResponse{}, nil
}

func (s *AuthServer) RefreshTokens(ctx context.Context, req *authv1.RefreshTokensRequest) (*authv1.RefreshTokensResponse, error) {
	ctx = authctx.WithUserGUID(ctx, req.GetUserGuid())
	ctx = authctx.WithSessionID(ctx, req.GetSessionId())

	tokenPair, err := s.auth.RefreshTokens(
		ctx,
		req.GetRefreshToken(),
		req.GetUserAgent(),
		req.GetIp(),
	)
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.RefreshTokensResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, services.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, services.ErrInvalidToken):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, services.ErrInvalidCode):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, services.ErrAccessDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, services.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, services.ErrUserExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, services.ErrOrganizationExists):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
