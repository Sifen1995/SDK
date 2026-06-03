package infrastructure

import (
	"context"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type goRedisAdapter struct {
	client *redis.Client
}

func (a *goRedisAdapter) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return a.client.SetNX(ctx, key, value, expiration).Result()
}

// NewRedisClientFromAddr builds a go-redis client when addr is configured.
func NewRedisClientFromAddr(addr string) redisClient {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return nil
	}
	return &goRedisAdapter{client: redis.NewClient(&redis.Options{Addr: addr})}
}

// noop for tests without redis
type noopRedis struct{}

func (noopRedis) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return true, nil
}
