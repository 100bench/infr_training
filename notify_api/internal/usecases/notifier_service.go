package usecases

import (
	"context"
	"fmt"
	
	er "github.com/100bench/infr_training/pkg/errors"
	en "github.com/100bench/infr_training/notify_api/internal/entities"
)


type NotifierService struct {
	notifier Notifier
}

func NewNotifierService(notifier Notifier) (*NotifierService, error) {
	if notifier == nil {
		return nil, fmt.Errorf("%w: NewNotifierService()", er.ErrNilDependency)
	}
	return &NotifierService{notifier: notifier}, nil
}

func (s *NotifierService) SendNotify(ctx context.Context, eventType en.NotificationType, id, title string) error {
	payLoad:= map[string]any{"id": id, "title": title}
	notification := en.NewNotification(eventType, payLoad)
	err := s.notifier.Send(ctx, notification)
	if err != nil {
		return fmt.Errorf("notifier service SendNotify(): %w", err)
	}
	return nil
}