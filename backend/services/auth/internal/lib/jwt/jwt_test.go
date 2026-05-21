package jwt

import (
	"errors"
	"testing"
	"time"

	"github.com/keshvan/rag-as-service/backend/services/auth/internal/entity"
)

func TestNewTokenPairAndParseToken(t *testing.T) {
	user := entity.User{
		GUID:           "user-1",
		OrganizationID: "org-1",
		Email:          "user@example.com",
	}

	tokenPair, err := NewTokenPairWithRefreshTTL(user, time.Minute, time.Hour, "secret-key", "session-1")
	if err != nil {
		t.Fatalf("NewTokenPairWithRefreshTTL returned error: %v", err)
	}

	if tokenPair.AccessToken == "" {
		t.Fatal("expected access token to be generated")
	}
	if tokenPair.RefreshToken == "" {
		t.Fatal("expected refresh token to be generated")
	}
	if tokenPair.RefreshTokenHash == "" {
		t.Fatal("expected refresh token hash to be generated")
	}
	if tokenPair.RefreshExpiresAt.IsZero() {
		t.Fatal("expected refresh expiry to be populated")
	}

	claims, err := ParseToken(tokenPair.AccessToken, "secret-key")
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}

	if claims.GUID != user.GUID {
		t.Fatalf("expected GUID %q, got %q", user.GUID, claims.GUID)
	}
	if claims.OrganizationID != user.OrganizationID {
		t.Fatalf("expected organization ID %q, got %q", user.OrganizationID, claims.OrganizationID)
	}
	if claims.SessionID != "session-1" {
		t.Fatalf("expected session ID %q, got %q", "session-1", claims.SessionID)
	}
	if claims.Type != "access" {
		t.Fatalf("expected access token type, got %q", claims.Type)
	}
}

func TestParseTokenAllowExpiredReturnsClaims(t *testing.T) {
	user := entity.User{
		GUID:           "user-1",
		OrganizationID: "org-1",
	}

	tokenPair, err := NewTokenPairWithRefreshTTL(user, -time.Minute, time.Hour, "secret-key", "session-1")
	if err != nil {
		t.Fatalf("NewTokenPairWithRefreshTTL returned error: %v", err)
	}

	_, err = ParseToken(tokenPair.AccessToken, "secret-key")
	if !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("expected ErrExpiredToken, got %v", err)
	}

	claims, err := ParseTokenAllowExpired(tokenPair.AccessToken, "secret-key")
	if err != nil {
		t.Fatalf("ParseTokenAllowExpired returned error: %v", err)
	}
	if claims.GUID != user.GUID || claims.OrganizationID != user.OrganizationID || claims.SessionID != "session-1" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestVerifyRefreshToken(t *testing.T) {
	user := entity.User{
		GUID:           "user-2",
		OrganizationID: "org-2",
	}

	tokenPair, err := NewTokenPair(user, time.Minute, "secret-key", "session-2")
	if err != nil {
		t.Fatalf("NewTokenPair returned error: %v", err)
	}

	if err := VerifyRefreshToken(tokenPair.RefreshToken, tokenPair.RefreshTokenHash); err != nil {
		t.Fatalf("VerifyRefreshToken returned error: %v", err)
	}
	if err := VerifyRefreshToken("wrong-token", tokenPair.RefreshTokenHash); err == nil {
		t.Fatal("expected VerifyRefreshToken to fail for invalid token")
	}
}
