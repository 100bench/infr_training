package cache

// старая реализация, теперь используется github.com/hashicorp/golang-lru/v2 в чкачестве лру ин меммори кеша

// import (
// 	"context"
// 	"sync"
// 	"time"
// )

// type item struct {
// 	value string
// 	expiration int64
// }

// type InMemmoryCache struct {
// 	items map[string]item
// 	mtx    sync.RWMutex
// 	interval time.Duration
// 	ttl time.Duration
// }

// func NewInMemmoryCache(ttl, interval time.Duration) *InMemmoryCache {
// 	return &InMemmoryCache{
// 		items: make(map[string]item),
// 		ttl: ttl,
// 		interval: interval,
// 	}
// }

// func (c *InMemmoryCache) Get(key string) (string, bool) {
// 	c.mtx.RLock()
// 	it, ok := c.items[key]
// 	c.mtx.RUnlock()
// 	if !ok{
// 		return "", false
// 	}

// 	if time.Now().UnixNano() > it.expiration {
// 		c.Delete(key)
// 		return "", false
// 	}
		
// 	return it.value, true
// }

// func (c *InMemmoryCache) Set(key string, value string) {
// 	c.mtx.Lock()
// 	defer c.mtx.Unlock()	
// 	c.items[key] = item{
// 		value: value,
// 		expiration: time.Now().Add(c.ttl).UnixNano(),
// 	}
// }

// func (c *InMemmoryCache) Delete(key string) {
// 	c.mtx.Lock()
// 	defer c.mtx.Unlock()
// 	delete(c.items, key)
// }

// func (c *InMemmoryCache) cleanUp() {
// 	now:= time.Now().UnixNano()
// 	c.mtx.Lock()
// 	defer c.mtx.Unlock()

// 	for k, v:= range c.items{
// 		if now > v.expiration{
// 			delete(c.items, k)
// 		}
// 	}

// }

// func (c *InMemmoryCache) Start(ctx context.Context){
// 	ticker := time.NewTicker(c.interval)

// 	go func() {
// 		defer ticker.Stop()

// 		for {
// 			select {
// 			case <-ctx.Done():
// 				return
// 			case <-ticker.C:
// 				c.cleanUp()
// 			}
// 		}
// 	}()
// }
