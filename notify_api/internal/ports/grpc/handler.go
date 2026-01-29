package grpc

import (
	"context"
	en "github.com/100bench/infr_training/notify_api/internal/entities"
)

type NotifierService interface {
	SendNotify(ctx context.Context, eventType en.NotificationType, id, title string) error
}