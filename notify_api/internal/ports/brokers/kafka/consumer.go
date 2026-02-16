package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"errors"

	en "github.com/100bench/infr_training/notify_api/internal/entities"
	er "github.com/100bench/infr_training/pkg/errors"
	"github.com/100bench/infr_training/pkg/logger"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/segmentio/kafka-go"
)



type KafkaConsumer struct {
	reader *kafka.Reader
	notifier NotifierService
	pool *pgxpool.Pool
	log logger.Logger
}

func NewKafkaConsumer(reader *kafka.Reader, notifier NotifierService, pool *pgxpool.Pool, log logger.Logger) (*KafkaConsumer, error){
	if reader == nil {
		return nil, fmt.Errorf("kafkaConsumer.Reader: %w", er.ErrNilDependency)
	}
	if notifier == nil {
		return nil, fmt.Errorf("kafkaConsumer.Notifier: %w", er.ErrNilDependency)
	}
	if pool == nil {
		return nil, fmt.Errorf("kafkaConsumer.Pool: %w", er.ErrNilDependency)
	}
	return &KafkaConsumer{
		reader: reader,
		notifier: notifier,
		pool: pool,
		log: log,
	}, nil
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	c.log.Info("kafka consumer started")
	for{
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg, err:= c.reader.FetchMessage(ctx)
		if err != nil{
			c.log.Error("Couldn't fetch message", logger.Field{"error", err})
			continue
		}

		err = c.processMessage(ctx, msg)
		if err != nil {
			c.log.Error("Couldn't process message", logger.Field{"error", err})
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.log.Error("Couldn't commit message", logger.Field{"error", err})
		}

	}
}

func (c *KafkaConsumer) processMessage(ctx context.Context, msg kafka.Message) error {
	eventID:= getHeader(msg.Headers, "event_id")
	eventType:= getHeader(msg.Headers, "event_type")
	if eventID == "" {
		return fmt.Errorf("missing event_id header")
	}
	var TaskEventPayload struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	err := json.Unmarshal(msg.Value, &TaskEventPayload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}
	
	tx, err:= c.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	query := `INSERT INTO processed_events (event_id, consumer_name) VALUES ($1, $2)`
	_, err = tx.Exec(ctx, query, eventID, "notify_api")
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == "23505" {
			c.log.Info("Event already processed, skipping", logger.Field{"event_id", eventID})
			tx.Rollback(ctx)
			return nil	
		}
		return fmt.Errorf("failed to insert processed event: %w", err)
	}

	switch eventType {
	case "task.created":
		err = c.notifier.SendNotify(ctx, en.TaskCreated, TaskEventPayload.ID, TaskEventPayload.Title)
		if err != nil {
			c.log.Error("failed to send notify, skipping", logger.Any("error", err))
		}
	case "task.deleted":
		err = c.notifier.SendNotify(ctx, en.TaskDeleted, TaskEventPayload.ID, TaskEventPayload.Title)
		if err != nil {
			c.log.Error("failed to send notify, skipping", logger.Any("error", err))
		}
	default:
		c.log.Warn("unknown event type, skipping", logger.Field{Key: "event_type", Value: eventType})
	}

	return tx.Commit(ctx)
}

func getHeader(headers []kafka.Header, key string) string {
	for _, h := range headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c *KafkaConsumer) Close() {
	if err := c.reader.Close(); err != nil {
		c.log.Error("failed to close kafka reader", logger.Field{"error", err})
	}
}