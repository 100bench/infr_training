package cache

import (
	"context"
	"time"
	"fmt"
	er "github.com/100bench/infr_training/pkg/errors"
	"github.com/100bench/infr_training/pkg/logger"
)

type cacheProvider interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Close()
}

type MetricsCache struct {
	cache cacheProvider
	log logger.Logger
}

func NewMetricsCache(cache cacheProvider, log logger.Logger) (*MetricsCache, error) {
	if cache == nil {
		return nil, fmt.Errorf("NewMetricsCache cache: %w", er.ErrNilDependency)
	}
	return &MetricsCache{
		cache: cache,
		log: log,
	}, nil
}

func (c *MetricsCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	err:= c.cache.Set(ctx, key, value, ttl)
	if err != nil{
		c.log.Error("Metrics.cache.Set(): ", logger.Field{"error", err})
		return fmt.Errorf("Metrics.cache.Set(): %w", err)
	}
	return nil
}

func (c *MetricsCache) Get(ctx context.Context, key string) (string, error) {
	value, err := c.cache.Get(ctx, key)
	if err != nil {
		cacheMissesTotal.Inc()
		c.log.Error("Metrics.cache.Get(): ", logger.Field{"error", err})
		return "", fmt.Errorf("MetricsCache.Get: %w", err)
	}
	return value, nil
}

func (c *MetricsCache) Delete(ctx context.Context, key string) error {
	err := c.cache.Delete(ctx, key)
	if err != nil {
		c.log.Error("Metrics.cache.Delete(): ", logger.Field{"error", err})
		return fmt.Errorf("MetricsCache.Delete: %w", err)
	}
	return nil
}

func (c *MetricsCache) Close() {
	c.cache.Close()
}