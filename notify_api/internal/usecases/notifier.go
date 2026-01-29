package usecases

import (
	"context"
	en "github.com/100bench/infr_training/notify_api/internal/entities"
)

type Notifier interface {
    Send(ctx context.Context, notification en.Notification) error
}