package jwt

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type GatewayClaims struct {
	UserGUID       string `json:"guid"`
	OrganizationID string `json:"org_id"`
	SessionID      string `json:"session_id"`
	Type           string `json:"type"`
	jwt.RegisteredClaims
}

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

func ParseToken(tokenString string, secret string) (*GatewayClaims, error) {
	return parseToken(tokenString, secret, false)
}

func ParseTokenAllowExpired(tokenString string, secret string) (*GatewayClaims, error) {
	return parseToken(tokenString, secret, true)
}

func parseToken(tokenString string, secret string, allowExpired bool) (*GatewayClaims, error) {
	parserOptions := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS512.Alg()}),
	}
	if allowExpired {
		parserOptions = append(parserOptions, jwt.WithoutClaimsValidation())
	}

	token, err := jwt.NewParser(parserOptions...).ParseWithClaims(tokenString, &GatewayClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*GatewayClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.Type != "access" ||
		claims.UserGUID == "" ||
		claims.OrganizationID == "" ||
		claims.SessionID == "" ||
		claims.ID == "" ||
		claims.ExpiresAt == nil {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
