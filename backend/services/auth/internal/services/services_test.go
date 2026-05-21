package services

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/keshvan/rag-as-service/backend/services/auth/internal/authctx"
	"github.com/keshvan/rag-as-service/backend/services/auth/internal/entity"
	"github.com/keshvan/rag-as-service/backend/services/auth/internal/lib/jwt"
)

type tokenRepoMock struct {
	saveTokenFunc               func(ctx context.Context, token *entity.RefreshToken) error
	getRefreshTokenByUserGUIDFn func(ctx context.Context, guid string) (*entity.RefreshToken, error)
	deleteTokenByUserGUIDFunc   func(ctx context.Context, guid string) error
}

func (m *tokenRepoMock) SaveToken(ctx context.Context, token *entity.RefreshToken) error {
	if m.saveTokenFunc != nil {
		return m.saveTokenFunc(ctx, token)
	}
	return nil
}

func (m *tokenRepoMock) GetRefreshTokenByUserGUID(ctx context.Context, guid string) (*entity.RefreshToken, error) {
	if m.getRefreshTokenByUserGUIDFn != nil {
		return m.getRefreshTokenByUserGUIDFn(ctx, guid)
	}
	return nil, nil
}

func (m *tokenRepoMock) DeleteTokenByUserGUID(ctx context.Context, guid string) error {
	if m.deleteTokenByUserGUIDFunc != nil {
		return m.deleteTokenByUserGUIDFunc(ctx, guid)
	}
	return nil
}

type userRepoMock struct {
	getUserByEmailFunc            func(ctx context.Context, email string) (entity.User, error)
	getUserByGUIDFunc             func(ctx context.Context, guid string) (entity.User, error)
	userExistsByEmailFunc         func(ctx context.Context, email string) (bool, error)
	createOrganizationWithOwnerFn func(ctx context.Context, organizationName string, organizationURL string, email string, passHash []byte, role string) (string, string, error)
	organizationExistsFunc        func(ctx context.Context, name, url string) (bool, error)
}

func (m *userRepoMock) GetUserByEmail(ctx context.Context, email string) (entity.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return entity.User{}, nil
}

func (m *userRepoMock) GetUserByGUID(ctx context.Context, guid string) (entity.User, error) {
	if m.getUserByGUIDFunc != nil {
		return m.getUserByGUIDFunc(ctx, guid)
	}
	return entity.User{}, nil
}

func (m *userRepoMock) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.userExistsByEmailFunc != nil {
		return m.userExistsByEmailFunc(ctx, email)
	}
	return false, nil
}

func (m *userRepoMock) CreateOrganizationWithOwner(
	ctx context.Context,
	organizationName string,
	organizationURL string,
	email string,
	passHash []byte,
	role string,
) (string, string, error) {
	if m.createOrganizationWithOwnerFn != nil {
		return m.createOrganizationWithOwnerFn(ctx, organizationName, organizationURL, email, passHash, role)
	}
	return "", "", nil
}

func (m *userRepoMock) OrganizationExists(ctx context.Context, name, url string) (bool, error) {
	if m.organizationExistsFunc != nil {
		return m.organizationExistsFunc(ctx, name, url)
	}
	return false, nil
}

type redisStorageMock struct{}

func (m *redisStorageMock) SaveCode(ctx context.Context, data entity.PendingUser, ttl time.Duration) error {
	return nil
}

func (m *redisStorageMock) GetCode(ctx context.Context, email string) (entity.PendingUser, error) {
	return entity.PendingUser{}, nil
}

func (m *redisStorageMock) DeleteCode(ctx context.Context, email string) error {
	return nil
}

type emailClientMock struct{}

func (m *emailClientMock) SendVerificationCode(to, code string) error {
	return nil
}

func TestRefreshTokensRotatesTokenForCurrentSession(t *testing.T) {
	user := entity.User{
		GUID:           "user-1",
		OrganizationID: "org-1",
		Email:          "user@example.com",
	}
	oldPair, err := jwt.NewTokenPairWithRefreshTTL(user, time.Minute, time.Hour, "secret-key", "session-1")
	if err != nil {
		t.Fatalf("NewTokenPairWithRefreshTTL returned error: %v", err)
	}

	var savedToken entity.RefreshToken
	auth := NewAuth(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		&tokenRepoMock{
			getRefreshTokenByUserGUIDFn: func(ctx context.Context, guid string) (*entity.RefreshToken, error) {
				if guid != user.GUID {
					t.Fatalf("expected guid %q, got %q", user.GUID, guid)
				}
				return &entity.RefreshToken{
					UserGUID:  user.GUID,
					TokenHash: oldPair.RefreshTokenHash,
					UserAgent: "browser",
					IP:        "127.0.0.1",
					SessionID: "session-1",
					ExpiresAt: time.Now().Add(time.Hour),
				}, nil
			},
			saveTokenFunc: func(ctx context.Context, token *entity.RefreshToken) error {
				savedToken = *token
				return nil
			},
		},
		&userRepoMock{
			getUserByGUIDFunc: func(ctx context.Context, guid string) (entity.User, error) {
				return user, nil
			},
		},
		&redisStorageMock{},
		&emailClientMock{},
		"secret-key",
		time.Minute,
		time.Hour,
	)

	ctx := authctx.WithUserGUID(context.Background(), user.GUID)
	ctx = authctx.WithSessionID(ctx, "session-1")

	newPair, err := auth.RefreshTokens(ctx, oldPair.RefreshToken, "browser", "127.0.0.2")
	if err != nil {
		t.Fatalf("RefreshTokens returned error: %v", err)
	}

	if newPair.AccessToken == "" || newPair.RefreshToken == "" {
		t.Fatal("expected rotated token pair")
	}
	if savedToken.UserGUID != user.GUID {
		t.Fatalf("expected saved guid %q, got %q", user.GUID, savedToken.UserGUID)
	}
	if savedToken.SessionID != "session-1" {
		t.Fatalf("expected same session to be preserved, got %q", savedToken.SessionID)
	}
	if savedToken.TokenHash == "" || savedToken.TokenHash == oldPair.RefreshTokenHash {
		t.Fatal("expected refresh token hash to be rotated")
	}
	if err := jwt.VerifyRefreshToken(newPair.RefreshToken, savedToken.TokenHash); err != nil {
		t.Fatalf("saved refresh hash does not match new refresh token: %v", err)
	}
}
