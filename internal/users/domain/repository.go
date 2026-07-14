package domain

import "context"

// UserRepository persists SDK users.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id int64) (*User, error)
	FindAll(ctx context.Context, limit, offset int) ([]*User, int64, error)
}
