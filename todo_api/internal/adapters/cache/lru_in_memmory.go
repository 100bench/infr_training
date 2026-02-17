package cache

import (
	"time"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

type LruMemCache struct {
	cache *expirable.LRU[string, string]
}

func NewLruMemCache(size int, ttl time.Duration) *LruMemCache {
	cache := expirable.NewLRU[string, string](size, nil, ttl)
	return &LruMemCache{
		cache: cache,
	}
}

func (c *LruMemCache) Get(key string) (V string, ok bool) {
	return c.cache.Get(key)
}

func (c *LruMemCache) Set(key string, value string) {
	c.cache.Add(key, value)
}

func (c *LruMemCache) Delete(key string) {
	c.cache.Remove(key)
}