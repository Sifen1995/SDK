package application

import (
	"context"
	"fmt"
	"time"

	"skykin-platform/internal/intents/domain"
	"skykin-platform/internal/intents/validator"
)

const activeIntentCacheTTL = 30 * time.Minute

// IntentService ingests on-device intent profiles and fetches a matching ad via ports.
type IntentService struct {
	profiles ProfileRepository
	cache    ActiveIntentCache
	ads      AdSelector
}

func NewIntentService(profiles ProfileRepository, cache ActiveIntentCache, ads AdSelector) *IntentService {
	return &IntentService{
		profiles: profiles,
		cache:    cache,
		ads:      ads,
	}
}

// IngestAndFetchAd validates the profile, caches it, persists it, and returns a campaign ad.
func (s *IntentService) IngestAndFetchAd(
	ctx context.Context,
	profile *domain.IntentProfile,
	channelCode string,
) (*IngestAndFetchAdResult, error) {
	if s == nil || s.profiles == nil || s.ads == nil {
		return nil, fmt.Errorf("intent service is not configured")
	}
	if err := validator.ValidateIntentProfile(profile); err != nil {
		return nil, err
	}

	normalizeProfileTimestamps(profile)

	if s.cache != nil {
		_ = s.cache.CacheActiveIntent(ctx, profile.PseudonymousID, profile.IntentName, activeIntentCacheTTL)
	}

	if err := s.profiles.Save(ctx, profile); err != nil {
		return nil, fmt.Errorf("save intent profile: %w", err)
	}

	ad, err := s.ads.SelectAd(ctx, profile.IntentName, channelCode)
	if err != nil {
		return nil, err
	}

	return &IngestAndFetchAdResult{
		PseudonymousID: profile.PseudonymousID,
		IntentName:     profile.IntentName,
		Confidence:     profile.Confidence,
		ModelVersion:   profile.ModelVersion,
		CampaignID:     ad.CampaignID,
		CampaignName:   ad.CampaignName,
		ChannelCode:    ad.ChannelCode,
		AdContent:      ad.Content,
	}, nil
}

func normalizeProfileTimestamps(profile *domain.IntentProfile) {
	now := time.Now().UTC()
	if profile.RecordedAt.IsZero() {
		profile.RecordedAt = now
	}
	if profile.ExpiresAt.IsZero() {
		profile.ExpiresAt = profile.RecordedAt.Add(24 * time.Hour)
	}
}
