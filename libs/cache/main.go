package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type Cache struct {
	redis *redis.Client
	ctx   context.Context
}

func NewCache(url string) Cache {
	return Cache{
		redis: redis.NewClient(&redis.Options{
			Addr:     url,
			Password: "",
			DB:       0,
		}),
		ctx: context.Background(),
	}
}

func (c *Cache) Get(key string) (string, error) {
	return c.redis.Get(c.ctx, key).Result()
}

func (c *Cache) SetWithTtl(key string, value string, ttl time.Duration) error {
	return c.redis.Set(c.ctx, key, value, ttl).Err()
}

func (c *Cache) SetKeepTtl(key string, value string) error {
	return c.redis.Do(c.ctx, "set", key, value, "keepttl").Err()
}

func (c *Cache) Delete(key string) {
	c.redis.Del(c.ctx, key)
}
