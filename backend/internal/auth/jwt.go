package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/goxm2/sports-league/internal/config"
)

// Claims is the access-token JWT payload.
type Claims struct {
	UserID   int64    `json:"uid"`
	Username string   `json:"usr"`
	Roles    []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}

// Manager issues and validates JWT access tokens and opaque refresh tokens.
type Manager struct {
	secret      []byte
	issuer      string
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

func NewManager(cfg config.JWTConfig) *Manager {
	return &Manager{
		secret:     []byte(cfg.Secret),
		issuer:     cfg.Issuer,
		accessTTL:  cfg.AccessTokenTTL,
		refreshTTL: cfg.RefreshTokenTTL,
	}
}

// AccessTTL exposes the configured access-token lifetime.
func (m *Manager) AccessTTL() time.Duration { return m.accessTTL }

// RefreshTTL exposes the configured refresh-token lifetime.
func (m *Manager) RefreshTTL() time.Duration { return m.refreshTTL }

// IssueAccessToken creates a signed JWT for the given user.
func (m *Manager) IssueAccessToken(userID int64, username string, roles []string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(m.secret)
}

// ParseAccessToken validates and decodes an access token.
func (m *Manager) ParseAccessToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// NewRefreshTokenID generates an opaque refresh token id (jti).
func NewRefreshTokenID() string {
	return uuid.NewString()
}

// RefreshExpiry returns the absolute expiry for a refresh token issued now.
func (m *Manager) RefreshExpiry() time.Time {
	return time.Now().Add(m.refreshTTL)
}
