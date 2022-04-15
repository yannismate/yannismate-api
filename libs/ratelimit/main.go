package ratelimit

import (
	log "github.com/sirupsen/logrus"
	"github.com/yannismate/yannismate-api/libs/cache"
	"strconv"
	"time"
)

type SharedRateLimiter struct {
	cache *cache.Cache
}

type LimitExceededError struct{}

func NewSharedRateLimiter(cache *cache.Cache) SharedRateLimiter {
	return SharedRateLimiter{
		cache: cache,
	}
}

func (srl *SharedRateLimiter) AllowIfTracked(key string) (int, error) {
	cacheRes, err := srl.cache.Get("ratelimiter:" + key)
	if err != nil {
		return 0, err
	}

	amountLeft, err := strconv.Atoi(cacheRes)
	if err != nil {
		log.WithField("event", "ratelimiter_parse").Error(err)
		return 0, err
	}

	amountLeft = amountLeft - 1

	if amountLeft < 0 {
		return amountLeft, nil
	}

	err = srl.cache.SetKeepTtl("ratelimiter:"+key, strconv.Itoa(amountLeft))
	if err != nil {
		log.WithField("event", "ratelimiter_update").Error(err)
		return 0, err
	}

	return amountLeft, nil
}

func (srl *SharedRateLimiter) AllowNew(key string, limit int, interval time.Duration) (int, error) {
	err := srl.cache.SetWithTtl("ratelimiter:"+key, strconv.Itoa(limit-1), interval)
	if err != nil {
		log.WithField("event", "ratelimiter_set").Error(err)
		return 0, err
	}
	return limit - 1, nil
}
