package app

import (
	"context"
	"fmt"
	// "net"
	"os"
	"os/signal"
	"syscall"
	"sync"

	// "google.golang.org/grpc"
	// "google.golang.org/grpc/reflection"

	// pb "github.com/100bench/infr_training/pkg/proto/gen/notification/v1"
	"github.com/100bench/infr_training/notify_api/internal/adapters/console"
	// grpcport "github.com/100bench/infr_training/notify_api/internal/ports/grpc" // alias, т.к. конфликт имён с google.golang.org/grpc
	"github.com/100bench/infr_training/notify_api/internal/usecases"
	kafkaport "github.com/100bench/infr_training/notify_api/internal/ports/brokers/kafka"
	log "github.com/100bench/infr_training/pkg/logger"
	"github.com/jackc/pgx/v4/pgxpool"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/100bench/infr_training/notify_api/config"
)

func RunApp() error {
	ctx, cancel := context.WithCancel(context.Background())

	logger, err := log.NewProductionLogger()
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	logger.Info("Starting Notification API...")

	cfg, err:= config.Load()
    if err != nil { 
        logger.Error("Failed to load configuration", log.Field{Key: "error", Value: err})
        return fmt.Errorf("%w: Failed to load configuration", err)
    }

	consoleNotifier, err := console.NewConsoleNotifier(ctx, logger)
	if err != nil {
		return fmt.Errorf("failed to create console notifier: %w", err)
	}
	pool, err := pgxpool.Connect(ctx, cfg.DSN)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer pool.Close()

	kafkaReader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   cfg.KafkaTopic,
		GroupID: "notify-api-group",
	})

	notifier, err := usecases.NewNotifierService(consoleNotifier)
	if err != nil {
		return fmt.Errorf("failed to create notifier service: %w", err)
	}
	
	kafkaConsumer, err := kafkaport.NewKafkaConsumer(kafkaReader, notifier, pool, logger)
	if err != nil {
		return fmt.Errorf("failed to create kafka consumer: %w", err)
	}


	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		kafkaConsumer.Start(ctx)
	}()


	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger.Info("Shutting down notification server...")
	cancel()
	wg.Wait()
	kafkaConsumer.Close()


	logger.Info("Notification server stopped")
	return nil
}
