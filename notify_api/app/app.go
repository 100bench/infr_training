package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/100bench/infr_training/pkg/proto/gen/notification/v1"

	"github.com/100bench/infr_training/notify_api/internal/adapters/console"
	grpcport "github.com/100bench/infr_training/notify_api/internal/ports/grpc" // alias, т.к. конфликт имён с google.golang.org/grpc
	"github.com/100bench/infr_training/notify_api/internal/usecases"
	log "github.com/100bench/infr_training/pkg/logger"
)

func RunApp() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := log.NewProductionLogger()
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	logger.Info("Starting Notification API...")


	consoleNotifier, err := console.NewConsoleNotifier(ctx, logger)
	if err != nil {
		return fmt.Errorf("failed to create console notifier: %w", err)
	}

	notifierService, err := usecases.NewNotifierService(consoleNotifier)
	if err != nil {
		return fmt.Errorf("failed to create notifier service: %w", err)
	}

	// === 5. СЛОЙ PORTS (gRPC) ===
	// Handler — реализует интерфейс pb.NotificationServiceServer.
	// Принимает gRPC запросы, конвертирует в вызовы NotifierService.
	handler, err := grpcport.NewHandler(notifierService)
	if err != nil {
		return fmt.Errorf("failed to create gRPC handler: %w", err)
	}

	// === 6. gRPC СЕРВЕР ===
	// grpc.NewServer() — создаёт новый gRPC сервер.
	// Можно передать опции: interceptors, credentials, etc.
	grpcServer := grpc.NewServer()

	// Регистрируем наш handler в gRPC сервере.
	// После этого сервер знает, что запросы к NotificationService
	// нужно направлять в наш handler.
	pb.RegisterNotificationServiceServer(grpcServer, handler)

	reflection.Register(grpcServer)

	// === 7. LISTENER ===
	// net.Listen создаёт TCP listener на указанном порту.
	// gRPC работает поверх TCP (или Unix sockets).
	port := ":50051"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", port, err)
	}

	go func() {
		logger.Info("gRPC server listening", log.String("port", port))
		if err := grpcServer.Serve(listener); err != nil {
			logger.Error("gRPC server error", log.Any("error", err))
			cancel()
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	logger.Info("Shutting down gRPC server...")

	grpcServer.GracefulStop()

	logger.Info("gRPC server stopped")
	return nil
}
