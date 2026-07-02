package application

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// PortalClaims is the minimal JWT payload for ad portal sessions.
type PortalClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (s *AuthService) signAccessToken(userID, role string, ttl time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().Add(ttl)
	claims := PortalClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := t.SignedString([]byte(s.cfg.JwtSecret))
	return token, expiresAt, err
}

func (s *AuthService) signRefreshToken(userID, role string) (string, error) {
	token, _, err := s.signAccessToken(userID, role, 7*24*time.Hour)
	return token, err
}
