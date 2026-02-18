package cache

import (
	"time"
	"context"
	"fmt"
	er "github.com/100bench/infr_training/pkg/errors"
	log "github.com/100bench/infr_training/pkg/logger"
)

type TwoLevelCache struct {
	level1  L1Cache
	level2 L2Cache
	logger log.Logger
}

func NewTwoLevelCache(l2cache L2Cache, l1cache L1Cache, logger log.Logger) (*TwoLevelCache, error) {
	if l2cache == nil {
		return nil, fmt.Errorf("NewTwoLevelCache l2cache: %w", er.ErrNilDependency)
	}
	if l1cache == nil {
		return nil, fmt.Errorf("NewTwoLevelCache l1cache: %w", er.ErrNilDependency)
	}
	return &TwoLevelCache{
		level1:  l1cache,
		level2: l2cache,
		logger: logger,
	}, nil
}

func (c *TwoLevelCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	c.level1.Set(key, value)

	if err := c.level2.Set(ctx, key, value, ttl); err != nil {
		c.logger.Error("TwoLevelCache.Set: ", log.Field{"error", err})
		return nil
	}
	return nil
}

func (c *TwoLevelCache) Get(ctx context.Context, key string) (string, error) {
	if value, ok := c.level1.Get(key); ok {
		cacheHitsTotal.WithLabelValues("l1").Inc()
		return value, nil
	}

	value, err := c.level2.Get(ctx, key)
	if err != nil {
		cacheL2ErrorsTotal.Inc()
		c.logger.Error("TwoLevelCache.Get: ", log.Field{"error", err})
		return "", fmt.Errorf("TwoLevelCache.Get: %w", err)	
	}
	cacheHitsTotal.WithLabelValues("l2").Inc()
	c.level1.Set(key, value)
	
	return value, nil
}

func (c *TwoLevelCache) Delete(ctx context.Context, key string) error {
	c.level1.Delete(key)
	if err := c.level2.Delete(ctx, key); err != nil {
		return fmt.Errorf("TwoLevelCache.Delete: %w", err)
	}
	return nil
}

func (c *TwoLevelCache) Close() {
	if err := c.level2.Close(); err != nil {
		fmt.Printf("TwoLevelCache.Close: failed to close level2 cache: %v\n", err)
	} 	
}