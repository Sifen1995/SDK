package application

import (
	"context"
	"errors"
	"time"

	frauddomain "skykin-platform/internal/fraud/domain"
)

var ErrFutureCursor = errors.New("since cursor cannot be in the future")

type SyncUseCase struct {
	repository frauddomain.SyncRepository
	now        func() time.Time
}

func NewSyncUseCase(repository frauddomain.SyncRepository) *SyncUseCase {
	return &SyncUseCase{
		repository: repository,
		now:        time.Now,
	}
}

// Execute returns a snapshot bounded by one server-generated cursor.
func (uc *SyncUseCase) Execute(
	ctx context.Context,
	since *time.Time,
) (*frauddomain.SyncSnapshot, error) {
	until := uc.now().UTC()
	if since != nil {
		normalized := since.UTC()
		if normalized.After(until) {
			return nil, ErrFutureCursor
		}
		since = &normalized
	}
	return uc.repository.Sync(ctx, since, until)
}
