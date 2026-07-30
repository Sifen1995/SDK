package application

import (
	"context"
	"time"

	"skykin-platform/internal/intents/domain"
)

// ProfileRepository persists SDK intent profiles (intents module only).
type ProfileRepository interface {
	Save(ctx context.Context, profile *domain.IntentProfile) error
}

// ActiveIntentCache caches the latest intent name per pseudonymous user (Job 3).
type ActiveIntentCache interface {
	CacheActiveIntent(ctx context.Context, pseudonymousID, intentName string, ttl time.Duration) error
}

// AdSelection is the campaigns module response DTO owned by the intents application boundary.
type AdSelection struct {
	CampaignID   string
	CampaignName string
	ChannelCode  string
	Content      map[string]any
}

// AdSelector selects a campaign ad for an intent without exposing campaign repositories here.
type AdSelector interface {
	SelectAd(ctx context.Context, pseudonymousID, targetIntent, channelCode string, smsConsented bool) (*AdSelection, error)
}

// SMSAdDispatcher dispatches an SMS+ campaign match (composition root wires delivery).
type SMSAdDispatcher interface {
	Dispatch(ctx context.Context, campaignID, pseudonymousID string) error
}
