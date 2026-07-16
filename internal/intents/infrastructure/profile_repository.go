package infrastructure

import (
	"context"

	intentdomain "skykin-platform/internal/intents/domain"
)

// ProfileRepository persists IntentProfile rows via the intents Postgres repository.
type ProfileRepository struct {
	intents intentdomain.IntentRepository
}

func NewProfileRepository(intents intentdomain.IntentRepository) *ProfileRepository {
	return &ProfileRepository{intents: intents}
}

func (r *ProfileRepository) Save(ctx context.Context, profile *intentdomain.IntentProfile) error {
	if r == nil || r.intents == nil {
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
