package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/100bench/infr_training/internal/adapters/storage"
	"github.com/100bench/infr_training/internal/ports/http/public"
	"github.com/100bench/infr_training/internal/usecases"
)

func RunApp() error {
    // 1. Dependencies
    storage := storage.NewInMemory()
    todoService := usecases.NewTodoService(storage)
    
    server, err := public.NewServer(todoService)
    if err != nil {
        return err
    }
    
    // 2. HTTP Server
    httpServer := &http.Server{
        Addr:    ":8080",
        Handler: server.GetRouter(),
    }
    
    // 3. Start server
    go func() {
        log.Println("Server starting on :8080")
        if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()
    
    // 4. Wait for signal
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
    <-stop
    
    log.Println("Shutting down server...")
    
    // 5. Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    return httpServer.Shutdown(ctx)
}