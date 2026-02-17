package cache

import (
	"context"
	"time"
)

type L2Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Close() error
}

type L1Cache interface {
	Get(key string) (string, bool)
	Set(key string, value string)
	Delete(key string)
}