package infrastructure

import (
	"context"
	"time"
)

// IntentCacheAdapter exposes RedisIntentRepository through the application cache port.
type IntentCacheAdapter struct {
	redis *RedisIntentRepository
}

func NewIntentCacheAdapter(redis *RedisIntentRepository) *IntentCacheAdapter {
	return &IntentCacheAdapter{redis: redis}
}

func (a *IntentCacheAdapter) CacheActiveIntent(
	ctx context.Context,
	pseudonymousID, intentName string,
	ttl time.Duration,
) error {
	if a == nil || a.redis == nil {
		return nil
	}
	return a.redis.CacheUserIntent(ctx, pseudonymousID, intentName, ttl)
}
