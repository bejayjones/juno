// Package auth provides JWT token issuance/verification and password hashing
// for the identity bounded context.
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims are the custom JWT claims stored in every Juno token.
type Claims struct {
	InspectorID string `json:"iid"`
	CompanyID   string `json:"cid"`
	Role        string `json:"role"`
	jwt.RegisteredClaims
}

// JWTService issues and verifies HS256 JWT tokens.
type JWTService struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTService(secret string, ttlHours int) *JWTService {
	return &JWTService{
		secret: []byte(secret),
		ttl:    time.Duration(ttlHours) * time.Hour,
	}
}

// Issue creates a signed token for the given inspector. Returns the token
// string and its expiry time.
func (s *JWTService) Issue(inspectorID, companyID, role string) (string, time.Time, error) {
	exp := time.Now().Add(s.ttl)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		InspectorID: inspectorID,
		CompanyID:   companyID,
		Role:        role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   inspectorID,
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, exp, nil
}

// Verify parses and validates a token, returning its claims on success.
func (s *JWTService) Verify(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
