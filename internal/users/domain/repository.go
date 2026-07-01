package domain

import "context"

// UserRepository resolves SDK users by external id.
type UserRepository interface {
	FindOrCreate(ctx context.Context, externalUserID string) (*User, error)
}
