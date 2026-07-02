package domain

import "context"

// UserRepository resolves SDK users by external id.
type UserRepository interface {
	FindOrCreate(ctx context.Context, externalUserID string) (*User, error)
	FindAll(ctx context.Context, limit, offset int) ([]*User, int64, error)
}
