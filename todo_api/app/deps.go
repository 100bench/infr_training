package app

import (
	"context"
	"fmt"

	log "github.com/100bench/infr_training/pkg/logger"
	"github.com/100bench/infr_training/todo_api/internal/adapters/cache"
	"github.com/100bench/infr_training/todo_api/internal/adapters/relay"
	"github.com/100bench/infr_training/todo_api/internal/adapters/storage/postgres"
	"github.com/100bench/infr_training/todo_api/internal/config"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/segmentio/kafka-go"
)

type deps struct {
	pool        *pgxpool.Pool
	storage     *postgres.TaskStorage
	cache       *cache.TwoLevelCache
	kafkaWriter *kafka.Writer
	relay       *relay.OutboxRelay
}

func initDeps(ctx context.Context, cfg *config.Config, logger log.Logger) (*deps, error) {
	pool, err := pgxpool.Connect(ctx, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	storage, err := postgres.NewTaskStorage(pool)
	if err != nil {
		return nil, fmt.Errorf("failed to create task storage: %w", err)
	}


	redisCache, err := cache.NewRedis(ctx, cfg.RedisAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}
	inMemmory := cache.NewLruMemCache(cfg.CacheSize, cfg.Ttl)
	twoLevelCache, err:= cache.NewTwoLevelCache(redisCache, inMemmory, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create two level cache: %w", err)
	}


	kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   cfg.KafkaTopic,
	})

	outboxRelay, err := relay.NewOutboxRelay(kafkaWriter, pool, cfg.RelayBatchSize, cfg.RelayPollInterval, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create outbox relay: %w", err)
	}

	return &deps{
		pool:        pool,
		storage:     storage,
		cache:       twoLevelCache,
		kafkaWriter: kafkaWriter,
		relay:       outboxRelay,
	}, nil
}

func (d *deps) close() {
	d.relay.Close()
	d.kafkaWriter.Close()
	d.cache.Close()
	d.pool.Close()
}
