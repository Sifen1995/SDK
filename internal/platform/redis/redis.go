// File: internal/platform/redis/redis.go
package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(addr string) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr, // e.g., "localhost:6379"
	})

	// Test if connection works
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &RedisClient{Client: rdb}, nil
}

// Del removes a key from Redis.
func (c *RedisClient) Del(ctx context.Context, key string) error {
	return c.Client.Del(ctx, key).Err()
}

// RPush pushes data onto the end of a Redis List (acting as our queue)
func (c *RedisClient) RPush(ctx context.Context, key string, value string) error {
	return c.Client.RPush(ctx, key, value).Err()
}

// BRPop blocks up to timeout waiting for a list item.
func (c *RedisClient) BRPop(ctx context.Context, key string, timeout time.Duration) (string, error) {
	results, err := c.Client.BRPop(ctx, timeout, key).Result()
	if err != nil {
		return "", err
	}
	if len(results) < 2 {
		return "", nil
	}
	return results[1], nil
}

// SetNX sets key only if it does not exist.
func (c *RedisClient) SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	return c.Client.SetNX(ctx, key, value, ttl).Result()
}

// LPush prepends values into a Redis list.
func (c *RedisClient) LPush(ctx context.Context, key string, values ...string) error {
	args := make([]interface{}, 0, len(values))
	for _, v := range values {
		args = append(args, v)
	}
	return c.Client.LPush(ctx, key, args...).Err()
}

// LRange reads list values from start to stop (inclusive).
func (c *RedisClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.Client.LRange(ctx, key, start, stop).Result()
}

// LTrim keeps only a selected range in a Redis list.
func (c *RedisClient) LTrim(ctx context.Context, key string, start, stop int64) error {
	return c.Client.LTrim(ctx, key, start, stop).Err()
}

// Expire sets key TTL.
func (c *RedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.Client.Expire(ctx, key, ttl).Err()
}

// ErrNil is returned by Get when the key does not exist.
var ErrNil = redis.Nil

// Set writes a key with an optional TTL (0 = no expiry).
func (c *RedisClient) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.Client.Set(ctx, key, value, ttl).Err()
}

// Get reads a string key. Returns ErrNil when the key is missing.
func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

// Incr atomically increments a key and returns the new value.
func (c *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return c.Client.Incr(ctx, key).Result()
}
