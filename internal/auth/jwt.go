// Package auth provides JWT token issuing and validation utilities.
// It wraps golang-jwt/jwt/v5 and exposes only what the CMS needs.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims embeds the standard JWT registered claims and adds CMS-specific fields.
type Claims struct {
	UserID int64  `json:"uid"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// TokenPair holds both access and refresh tokens.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// JWTService issues and validates tokens using a scoped secret.
// No global state — instantiate per application.
type JWTService struct {
	secret []byte
	expiry time.Duration
}

// NewJWTService creates a JWTService with the given HMAC secret and expiry.
func NewJWTService(secret string, expiry time.Duration) *JWTService {
	return &JWTService{
		secret: []byte(secret),
		expiry: expiry,
	}
}

// Issue creates a signed JWT for the given principal.
func (s *JWTService) Issue(userID int64, name, email, role string) (*TokenPair, error) {
	now := time.Now()
	expiresAt := now.Add(s.expiry)

	claims := Claims{
		UserID: userID,
		Name:   name,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Issuer:    "go-templio",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken: signed,
		ExpiresAt:   expiresAt,
	}, nil
}

// Validate parses and validates a token string, returning the embedded Claims.
func (s *JWTService) Validate(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("auth: unexpected signing method")
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("auth: invalid token claims")
	}

	return claims, nil
}
