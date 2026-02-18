package app

import (
	"context"
	"fmt"

	"github.com/100bench/infr_training/todo_api/internal/adapters/cache"
	"github.com/jackc/pgx/v4/pgxpool"
)

type appReadinessChecker struct {
    pool *pgxpool.Pool
    cache *cache.Redis
}

func (a *appReadinessChecker) Ready(ctx context.Context) error {
    if err := a.pool.Ping(ctx); err != nil {
        return fmt.Errorf("database not ready: %w", err)
    }
    
	if err:= a.cache.Ping(ctx); err != nil {
		return fmt.Errorf("redis not ready: %w", err)
	}
    return nil
}