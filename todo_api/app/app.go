package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cb "github.com/100bench/infr_training/pkg/circuit_breaker"
	log "github.com/100bench/infr_training/pkg/logger"
	"github.com/100bench/infr_training/todo_api/internal/adapters/grpc_client"
	"github.com/100bench/infr_training/todo_api/internal/config"
	"github.com/100bench/infr_training/todo_api/internal/ports/http/public"
	"github.com/100bench/infr_training/todo_api/internal/usecases"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)


func RunApp() error {
    logger, err:= log.NewDevelopmentLogger()
    if err != nil {
        return fmt.Errorf("%w: no zap logger", err)
    }
    
    logger.Info("Starting application...")

    cfg, err:= config.Load()
    if err != nil { 
        logger.Error("Failed to load configuration", log.Field{Key: "error", Value: err})
        return fmt.Errorf("%w: Failed to load configuration", err)
    }

    ctx := context.Background()

    // 1. Dependencies
    d, err := initDeps(ctx, cfg, logger)
    if err != nil {
        logger.Error("Failed to init dependencies", log.Field{Key: "error", Value: err})
        return fmt.Errorf("%w: Failed to init dependencies", err)
    }
    defer d.close()

    // 2. gRPC notifier (временно, до полного перехода на Kafka)
    conn, err := grpc.NewClient(cfg.NotifyServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        logger.Error("Failed to connect to notification service", log.Field{Key: "error", Value: err})
        return fmt.Errorf("%w: Failed to connect to notification service", err)
    }
    defer conn.Close()

    notificationClient := grpc_client.NewNotificationClient(conn)
    circuitBreaker := *cb.NewCircuitBreaker(5, 3, 2, 30*time.Second)
    notifierWithCB := grpc_client.NewNotifierWithCircuitBreaker(notificationClient, &circuitBreaker)

    // 3. Start relay
    relayCtx, relayCancel := context.WithCancel(ctx)
    defer relayCancel()
    go d.relay.Start(relayCtx)

    todoService, err := usecases.NewTodoService(ctx, d.storage, d.cache, notifierWithCB, logger)
    if err != nil {
        var field log.Field
        field = log.Field{Key: "error", Value: err}
        logger.Error("Failed to create todo service", field)
        return fmt.Errorf("%w: Failed to create todo service", err)
    }

    server, err := public.NewServer(todoService)
    if err != nil {
        return err
    }
    
    // 2. HTTP Server
    httpServer := &http.Server{
        Addr:    ":8080",
        Handler: server.GetRouter(),
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    // 3. Start server
    go func() {
        logger.Info("Server starting on :8080")
        if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            var field log.Field
            field = log.Field{Key: "error", Value: err}
            logger.Error("Server error", field)
        }
    }()
    
    // 4. Wait for signal
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
    <-stop
    
    logger.Info("Shutting down server...")
    
    // 5. Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    return httpServer.Shutdown(ctx)
}