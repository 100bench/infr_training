package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/100bench/infr_training/internal/adapters/cache"
	"github.com/100bench/infr_training/internal/adapters/storage/postgres"
	"github.com/100bench/infr_training/internal/ports/http/public"
	"github.com/100bench/infr_training/internal/usecases"
	"go.uber.org/zap"
)

func RunApp() error {
    logger, err:= zap.NewProduction()
    if err != nil {
        return fmt.Errorf("%w: no zap logger", err)
    }

    defer logger.Sync()
    sugar:= logger.Sugar()
    sugar.Info("Starting application...")

    dsn:= os.Getenv("DSN_STRING")
    if dsn == "" {
        sugar.Errorw("DSN_STRING environment variable is not set")
    }
    addr:= os.Getenv("REDIS_ADDR")

    ctx:= context.Background()

    // 1. Dependencies
    storage, err := postgres.NewTaskStorage(ctx, dsn)
    if err != nil {
        sugar.Errorw("Failed to create task storage:", "error", err)
    }
    defer storage.Close()

    redisCache, err := cache.NewRedisCache(ctx, addr)
    if err != nil {
        sugar.Errorw("Failed to create cache:", "error", err)
    }
    defer redisCache.Close()
    
    todoService, err := usecases.NewTodoService(storage, redisCache)
    if err != nil {
        sugar.Errorw("Failed to create todo service:", "error", err)
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
            sugar.Errorw("Server error", "error", err)
        }
    }()
    
    // 4. Wait for signal
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
    <-stop
    
    sugar.Infow("Shutting down server...")
    
    // 5. Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    return httpServer.Shutdown(ctx)
}