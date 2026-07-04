package infrastructure

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const cacheKeyPrefix = "perm_cache:"

type RedisPermissionCache struct {
	rdb    *redis.Client
	logger *slog.Logger
}

func NewRedisPermissionCache(rdb *redis.Client, logger *slog.Logger) *RedisPermissionCache {
	if logger == nil {
		logger = slog.Default()
	}
	return &RedisPermissionCache{rdb: rdb, logger: logger}
}

func (c *RedisPermissionCache) Get(roleName string) ([]string, bool) {
	val, err := c.rdb.Get(context.Background(), cacheKeyPrefix+roleName).Result()
	if err != nil {
		return nil, false
	}
	var perms []string
	if err := json.Unmarshal([]byte(val), &perms); err != nil {
		return nil, false
	}
	return perms, true
}

func (c *RedisPermissionCache) Set(roleName string, permissions []string, ttl time.Duration) {
	data, err := json.Marshal(permissions)
	if err != nil {
		c.logger.Warn("permission cache marshal failed", "role", roleName, "error", err)
		return
	}
	if err := c.rdb.Set(context.Background(), cacheKeyPrefix+roleName, data, ttl).Err(); err != nil {
		c.logger.Warn("permission cache set failed", "role", roleName, "error", err)
	}
}

func (c *RedisPermissionCache) Invalidate(roleName string) {
	if err := c.rdb.Del(context.Background(), cacheKeyPrefix+roleName).Err(); err != nil {
		c.logger.Warn("permission cache invalidate failed", "role", roleName, "error", err)
	}
}

type inMemoryEntry struct {
	perms []string
}

type InMemoryPermissionCache struct {
	store sync.Map
}

func NewInMemoryPermissionCache() *InMemoryPermissionCache {
	return &InMemoryPermissionCache{}
}

func (c *InMemoryPermissionCache) Get(roleName string) ([]string, bool) {
	val, ok := c.store.Load(roleName)
	if !ok {
		return nil, false
	}
	entry, _ := val.(inMemoryEntry)
	return entry.perms, true
}

func (c *InMemoryPermissionCache) Set(roleName string, permissions []string, _ time.Duration) {
	c.store.Store(roleName, inMemoryEntry{perms: permissions})
}

func (c *InMemoryPermissionCache) Invalidate(roleName string) {
	c.store.Delete(roleName)
}
