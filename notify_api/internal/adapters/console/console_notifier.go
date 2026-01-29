package console

import (
	"context"
	"fmt"

	en "github.com/100bench/infr_training/notify_api/internal/entities"
	er "github.com/100bench/infr_training/pkg/errors"
	log "github.com/100bench/infr_training/pkg/logger"
)

type ConsoleNotifier struct {
	logger log.Logger
}

func NewConsoleNotifier(ctx context.Context, logger log.Logger) (*ConsoleNotifier, error) {
	if logger == nil {
		return nil, fmt.Errorf("%w: NewConsoleNotifier()", er.ErrNilDependency)
	}
	return &ConsoleNotifier{logger: logger}, nil
}

func (s *ConsoleNotifier) Send(ctx context.Context, notification en.Notification) error {
	if s.logger == nil {
		return fmt.Errorf("%w: ConsoleNotifier.Send()", er.ErrNilDependency)
	}

	logger := s.logger.WithContext(ctx)

	switch notification.Type {
	case en.TaskCreated:
		logger.Info("task created", log.Any("payload", notification.Payload))
	case en.TaskDeleted:
		logger.Info("task deleted", log.Any("payload", notification.Payload))
	default:
		return fmt.Errorf("unknown notification type: %d", notification.Type)
	}

	return nil
}
