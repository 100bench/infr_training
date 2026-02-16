package relay

import (
	"context"
	"fmt"
	"time"

	"github.com/100bench/infr_training/pkg/logger"

	er "github.com/100bench/infr_training/pkg/errors"
	en "github.com/100bench/infr_training/todo_api/internal/entities"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/segmentio/kafka-go"
)

type OutboxRelay struct{
	writer *kafka.Writer
	pool *pgxpool.Pool
	log logger.Logger
	batchSize int
	pollInterval time.Duration
}

func NewOutboxRelay(writer *kafka.Writer, pool *pgxpool.Pool, batchSize int, pollInterval time.Duration, log logger.Logger) (*OutboxRelay, error) {
	if writer == nil {
		return nil, fmt.Errorf("outboxRelay.Writer: %w", er.ErrNilDependency)
	}
	if pool == nil {
		return nil, fmt.Errorf("outboxRelay.Pool: %w", er.ErrNilDependency)
	}
	return &OutboxRelay{
		writer: writer,
		pool: pool,
		log: log,
		batchSize: batchSize,
		pollInterval: pollInterval,
	}, nil
}

func (r *OutboxRelay) Start(ctx context.Context) {
	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()
	r.log.Info("outbox relay started", logger.Int("batch_size", r.batchSize))
	for {
		select {
			case <-ticker.C:
				if err := r.relayBatch(ctx); err != nil {
					r.log.Error("Error relaying batch", logger.Any("error", err))
				}
			case <-ctx.Done():
				return
		}
	}
}

func (r *OutboxRelay) relayBatch(ctx context.Context) error {
	tx, err:= r.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil{
		return fmt.Errorf("outbxoxRelay.relayBatch(): %w", err)
	}
	defer tx.Rollback(ctx)

	const q = `SELECT id, task_id, event_type, payload, sent, created_at FROM outbox WHERE sent = false ORDER BY created_at 
	LIMIT $1 FOR UPDATE SKIP LOCKED`
	rows, err := tx.Query(ctx, q, r.batchSize)
	if err != nil {
		return fmt.Errorf("outboxRelay.relayBatch(): %w", err)
	}
	defer rows.Close()


	var events []en.OutboxEvent
	for rows.Next() {
		var e en.OutboxEvent
		if err := rows.Scan(&e.ID, &e.TaskID, &e.EventType, &e.Payload, &e.Sent, &e.CreatedAt); err != nil {
			return fmt.Errorf("outboxRelay.relayBatch(): %w", err)
		}
		events = append(events, e)
	}

	if len(events) == 0 {
		return tx.Commit(ctx)
	}
	
	// Метрика размера батча
	outboxBatchSize.Observe(float64(len(events)))

	// Метрика lag (старейшее событие)
	oldestAge := time.Since(events[0].CreatedAt).Seconds()
	outboxLagSeconds.Set(oldestAge)


	// отпаврка во внешний брокер
	// не уверен, так ли надо отправлять в нашем случае..
	// если мы используем тут только айди и payload то зачем сканили остальные поля?
	// жду ответа на ревью
	msgs:= make([]kafka.Message, len(events))
	for i, e := range events {
		msgs[i] = kafka.Message{
			Key: []byte(e.TaskID.String()),
			Value: e.Payload,
			Headers: []kafka.Header{
				{Key: "event_type", Value: []byte(e.EventType)},
				{Key: "event_id", Value: []byte(e.ID.String())},
			},
		}
	}
	if err := r.writer.WriteMessages(ctx, msgs...); err != nil {
		return fmt.Errorf("outboxRelay.relayBatch(): %w", err)
	}

	ids:= make([]uuid.UUID, 0, len(events))
	for _, e:= range events{
		ids = append(ids, e.ID)
	}

	const q2 = `UPDATE outbox SET sent = true WHERE id = ANY($1)`
	if _, err = tx.Exec(ctx, q2, ids); err != nil{
		outboxErrors.WithLabelValues("database").Inc()

		return fmt.Errorf("outboxRelay.relayBatch(): %w", err)
	}


	outboxEventsProcessed.Add(float64(len(events)))

	r.log.Info("batch relayed",
		logger.Int("count", len(events)),
		logger.Float64("lag_seconds", oldestAge),
	)

	return tx.Commit(ctx)
}

func (r *OutboxRelay) Close() error {
	return r.writer.Close()
}