package infrastructure

import (
	"context"

	intentdomain "skykin-platform/internal/intents/domain"
)

// ProfileRepository persists intent profiles via async queue or direct Postgres fallback.
type ProfileRepository struct {
	intents intentdomain.IntentRepository
	queue   *IntentLogQueue
}

func NewProfileRepository(intents intentdomain.IntentRepository, queue *IntentLogQueue) *ProfileRepository {
	return &ProfileRepository{intents: intents, queue: queue}
}

func (r *ProfileRepository) Save(ctx context.Context, profile *intentdomain.IntentProfile) error {
	if r == nil {
		return nil
	}
	if r.queue != nil {
		return r.queue.Enqueue(ctx, profile)
	}
	if r.intents == nil {
		return nil
	}
	intent := &intentdomain.Intent{
		UserID:     profile.PseudonymousID,
		IntentName: profile.IntentName,
		Confidence: profile.Confidence,
		CreatedAt:  profile.RecordedAt,
	}
	_, err := r.intents.Create(ctx, intent)
	return err
}
