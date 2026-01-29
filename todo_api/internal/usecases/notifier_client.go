package usecases

import (
	"context"
	pb "github.com/100bench/infr_training/pkg/proto/gen/notification/v1"
)

type Notifier interface{
	Send(ctx context.Context, id, title string, notifyType pb.EventType) error
}