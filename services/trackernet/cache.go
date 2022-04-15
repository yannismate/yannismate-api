package main

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
