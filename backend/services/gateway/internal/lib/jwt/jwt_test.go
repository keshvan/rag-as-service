package jwt

import (
	"errors"
	"testing"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func signedTestToken(t *testing.T, expiresAt time.Time) string {
	t.Helper()

	claims := GatewayClaims{
		UserGUID:       "user-1",
		OrganizationID: "org-1",
		SessionID:      "session-1",
		Type:           "access",
		RegisteredClaims: jwtv5.RegisteredClaims{
			ID:        "token-1",
			Subject:   "user-1",
			ExpiresAt: jwtv5.NewNumericDate(expiresAt),
			IssuedAt:  jwtv5.NewNumericDate(time.Now()),
			NotBefore: jwtv5.NewNumericDate(time.Now()),
		},
	}

	token, err := jwtv5.NewWithClaims(jwtv5.SigningMethodHS512, claims).SignedString([]byte("secret-key"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	return token
}

func TestParseToken(t *testing.T) {
	token := signedTestToken(t, time.Now().Add(time.Minute))

	claims, err := ParseToken(token, "secret-key")
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}

	if claims.UserGUID != "user-1" || claims.OrganizationID != "org-1" || claims.SessionID != "session-1" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestParseTokenAllowExpired(t *testing.T) {
	token := signedTestToken(t, time.Now().Add(-time.Minute))

	_, err := ParseToken(token, "secret-key")
	if !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("expected ErrExpiredToken, got %v", err)
	}

	claims, err := ParseTokenAllowExpired(token, "secret-key")
	if err != nil {
		t.Fatalf("ParseTokenAllowExpired returned error: %v", err)
	}
	if claims.UserGUID != "user-1" || claims.SessionID != "session-1" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}
