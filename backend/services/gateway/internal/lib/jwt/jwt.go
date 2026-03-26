package jwt

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type GatewayClaims struct {
	UserGUID       string `json:"guid"`
	OrganizationID string `json:"org_id"`
	SessionID      string `json:"session_id"`
	jwt.RegisteredClaims
}

var ErrInvalidToken = errors.New("invalid token")

func ParseToken(tokenString string, secret string) (*GatewayClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &GatewayClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*GatewayClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
