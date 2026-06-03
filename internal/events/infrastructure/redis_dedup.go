package infrastructure

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const dedupTTL = 5 * time.Minute

// RedisDedupStore deduplicates event IDs using Redis SET NX with TTL.
// Falls back to in-memory dedup when Redis is unavailable.
type RedisDedupStore struct {
	mu    sync.Mutex
	mem   map[string]time.Time
	redis redisClient
}

type redisClient interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
}

func NewRedisDedupStore(client redisClient) *RedisDedupStore {
	return &RedisDedupStore{
		mem:   make(map[string]time.Time),
		redis: client,
	}
}

func (s *RedisDedupStore) TryAcquire(ctx context.Context, eventID string) (bool, error) {
	key := "event:dedup:" + eventID

	if s.redis != nil {
		ok, err := s.redis.SetNX(ctx, key, "1", dedupTTL)
		if err != nil {
			return false, fmt.Errorf("redis dedup: %w", err)
		}
		return ok, nil
	}

	return s.tryAcquireMemory(key), nil
}

func (s *RedisDedupStore) tryAcquireMemory(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for k, expires := range s.mem {
		if now.After(expires) {
			delete(s.mem, k)
		}
	}
	if _, exists := s.mem[key]; exists {
		return false
	}
	s.mem[key] = now.Add(dedupTTL)
	return true
}
