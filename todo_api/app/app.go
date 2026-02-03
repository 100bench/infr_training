package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/100bench/infr_training/todo_api/internal/adapters/cache"
	"github.com/100bench/infr_training/todo_api/internal/adapters/storage/postgres"
	"github.com/100bench/infr_training/todo_api/internal/ports/http/public"
	"github.com/100bench/infr_training/todo_api/internal/usecases"
    log "github.com/100bench/infr_training/pkg/logger"
    "github.com/100bench/infr_training/todo_api/internal/adapters/grpc_client"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    cb "github.com/100bench/infr_training/pkg/circuit_breaker"
)


func RunApp() error {
    logger, err:= log.NewDevelopmentLogger()
    if err != nil {
        return fmt.Errorf("%w: no zap logger", err)
    }
    
    logger.Info("Starting application...")

    dsn:= os.Getenv("DSN_STRING")
    if dsn == "" {
        return fmt.Errorf("%w: DSN_STRING environment variable is not set", err)
    }

    addr:= os.Getenv("REDIS_ADDR")
    if addr == "" {
        return fmt.Errorf("%w: REDIS_ADDR environment variable is not set", err)
    }

    ctx:= context.Background()

    // 1. Dependencies
    storage, err := postgres.NewTaskStorage(ctx, dsn)
    if err != nil {
        var field log.Field
        field = log.Field{Key: "error", Value: err}
        logger.Error("Failed to create task storage", field)
        return fmt.Errorf("%w: Failed to create task storage", err)
        
    }
    defer storage.Close()

    redisCache, err := cache.NewRedisCache(ctx, addr)
    if err != nil {
        var field log.Field
        field = log.Field{Key: "error", Value: err}
        logger.Error("Failed to create cache", field)
        return fmt.Errorf("%w: Failed to create cache", err)
    }
    defer redisCache.Close()

    notifyAddr:= os.Getenv("NOTIFY_SERVICE_ADDR")
    if notifyAddr == "" {
        return fmt.Errorf("%w: NOTIFY_SERVICE_ADDR environment variable is not set", err)
    }
    conn, err := grpc.NewClient(notifyAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        var field log.Field
        field = log.Field{Key: "error", Value: err}
        logger.Error("Failed to connect to notification service", field)
        return fmt.Errorf("%w: Failed to connect to notification service", err)
    }
    defer conn.Close()  

    notificationClient := grpc_client.NewNotificationClient(conn)
    cb:= *cb.NewCircuitBreaker(5, 3, 2, 30*time.Second)
    notifierWithCB:= grpc_client.NewNotifierWithCircuitBreaker(notificationClient, &cb)

    
    todoService, err := usecases.NewTodoService(storage, redisCache, notifierWithCB, logger)
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