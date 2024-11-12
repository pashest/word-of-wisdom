package cache

import (
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/pashest/word-of-wisdom/internal/model"
)

type RequestCache struct {
	cache      *ttlcache.Cache[string, *model.Challenge]
	defaultTTL time.Duration
}

func NewCache(defaultTTL time.Duration) *RequestCache {
	return &RequestCache{
		defaultTTL: defaultTTL,
		cache:      ttlcache.New[string, *model.Challenge](ttlcache.WithTTL[string, *model.Challenge](defaultTTL)),
	}
}

func (c *RequestCache) Get(key string) (*model.Challenge, bool) {
	if !c.cache.Has(key) {
		return nil, false
	}
	return c.cache.Get(key).Value(), true
}

func (c *RequestCache) Set(key string, challenge *model.Challenge) {
	c.cache.Set(key, challenge, c.defaultTTL)
}

func (c *RequestCache) Delete(key string) {
	c.cache.Delete(key)
}
