package application

import (
	"context"
	"errors"
	"time"
)

var errInvalidCredentials = errors.New("invalid credentials")

// ErrInvalidCredentials is returned for all login failures (exported for handlers).
var ErrInvalidCredentials = errInvalidCredentials

func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	u, lookupErr := s.repo.GetPortalUserByEmail(ctx, email)

	hashToCompare := loginDummyHash
	active := false
	var userID, role string
	if lookupErr == nil && u != nil {
		hashToCompare = u.PasswordHash
		active = u.IsActive
		userID = u.ID
		role = u.RoleSlug()
	}

	if !bcryptPasswordMatches(hashToCompare, password) || lookupErr != nil || !active {
		return nil, errInvalidCredentials
	}

	token, expiresAt, err := s.signAccessToken(userID, role, accessTokenTTL)
	if err != nil {
		return nil, err
	}
	refresh, err := s.signRefreshToken(userID, role)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Token:        token,
		RefreshToken: refresh,
		ExpiresAt:    expiresAt.Format(time.RFC3339),
		User:         UserInfo{ID: userID, Role: role},
	}, nil
}
