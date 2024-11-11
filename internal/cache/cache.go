package cache

import (
	"time"

	"github.com/jellydator/ttlcache/v3"
)

type RequestCache struct {
	cache      *ttlcache.Cache[string, struct{}]
	defaultTTL time.Duration
}

func NewCache(defaultTTL time.Duration) *RequestCache {
	return &RequestCache{
		defaultTTL: defaultTTL,
		cache:      ttlcache.New[string, struct{}](ttlcache.WithTTL[string, struct{}](defaultTTL)),
	}
}

func (c *RequestCache) Get(key string) bool {
	return c.cache.Has(key)
}

func (c *RequestCache) Set(key string) {
	c.cache.Set(key, struct{}{}, c.defaultTTL)
}

func (c *RequestCache) Delete(key string) {
	c.cache.Delete(key)
}
