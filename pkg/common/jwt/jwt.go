package jwt

import (
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims.
type Claims struct {
	gojwt.RegisteredClaims
	UserID uint64 `json:"user_id"`
	Role   string `json:"role"`
}

// TokenPair contains access and refresh tokens.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Manager handles JWT token operations.
type Manager struct {
	secret             []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

// NewManager creates a new JWT manager.
func NewManager(secret string, accessExpiry, refreshExpiry time.Duration) *Manager {
	return &Manager{
		secret:             []byte(secret),
		accessTokenExpiry:  accessExpiry,
		refreshTokenExpiry: refreshExpiry,
	}
}

// GenerateTokenPair creates a new access/refresh token pair.
func (m *Manager) GenerateTokenPair(userID uint64, role string) (*TokenPair, error) {
	now := time.Now()

	accessClaims := Claims{
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(now.Add(m.accessTokenExpiry)),
			IssuedAt:  gojwt.NewNumericDate(now),
			Subject:   fmt.Sprintf("%d", userID),
		},
		UserID: userID,
		Role:   role,
	}
	accessToken := gojwt.NewWithClaims(gojwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString(m.secret)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshClaims := Claims{
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(now.Add(m.refreshTokenExpiry)),
			IssuedAt:  gojwt.NewNumericDate(now),
			Subject:   fmt.Sprintf("%d", userID),
		},
		UserID: userID,
		Role:   role,
	}
	refreshToken := gojwt.NewWithClaims(gojwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString(m.secret)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		ExpiresIn:    int64(m.accessTokenExpiry.Seconds()),
	}, nil
}

// ValidateToken parses and validates a JWT token string.
func (m *Manager) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := gojwt.ParseWithClaims(tokenStr, &Claims{}, func(t *gojwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
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

// RefreshTokenExpiry returns the refresh token expiry duration.
func (m *Manager) RefreshTokenExpiry() time.Duration {
	return m.refreshTokenExpiry
}
