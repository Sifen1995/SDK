package infrastructure

import (
	"context"

	"skykin-platform/internal/audience/domain"

	"gorm.io/gorm"
)

// UnitOfWork binds the audience repositories to a single database transaction so
// segment publication (segment row, memberships, candidate status) is all-or-nothing.
type UnitOfWork struct {
	db *gorm.DB
}

func NewUnitOfWork(db *gorm.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

var _ domain.UnitOfWork = (*UnitOfWork)(nil)

func (u *UnitOfWork) Do(ctx context.Context, fn func(r domain.Repositories) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(domain.Repositories{
			Segments:   NewSegmentRepository(tx),
			Membership: NewMembershipRepository(tx),
			Candidates: NewCandidateRepository(tx),
		})
	})
}
