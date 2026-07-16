package infrastructure

import (
	"context"
	"fmt"
	"time"

	"skykin-platform/internal/platform/redis"
)

// RedisIntentRepository caches active user intents via the platform Redis client.
type RedisIntentRepository struct {
	rdb *redis.RedisClient
}

func NewRedisIntentRepository(rdb *redis.RedisClient) *RedisIntentRepository {
	return &RedisIntentRepository{rdb: rdb}
}

// CacheUserIntent stores the active user profile intent name (Job 3).
func (r *RedisIntentRepository) CacheUserIntent(ctx context.Context, pseudonymousID string, intent string, ttl time.Duration) error {
	if r == nil || r.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("user_intent:%s", pseudonymousID)
	return r.rdb.Set(ctx, key, intent, ttl)
}

// GetUserIntent retrieves the cached active intent.
func (r *RedisIntentRepository) GetUserIntent(ctx context.Context, pseudonymousID string) (string, error) {
	if r == nil || r.rdb == nil {
		return "", nil
	}
	key := fmt.Sprintf("user_intent:%s", pseudonymousID)
	val, err := r.rdb.Get(ctx, key)
	if err == redis.ErrNil {
		return "", nil // cache miss
	}
	return val, err
}
