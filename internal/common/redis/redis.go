package redis

// RedisClient is a placeholder for redis cache connections.
type RedisClient struct {
	Addr string
}

func NewRedisClient(addr string) *RedisClient {
	return &RedisClient{Addr: addr}
}
