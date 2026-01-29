package grpc

import (
	pb "github.com/100bench/infr_training/pkg/proto/gen/notification/v1"
	er "github.com/100bench/infr_training/pkg/errors"
	en "github.com/100bench/infr_training/notify_api/internal/entities"
	"fmt"
	"context"
)

type Handler struct {
	pb.UnimplementedNotificationServiceServer
	service NotifierService
}

func NewHandler(service NotifierService) (*Handler, error) {
	if service == nil {
		return nil, fmt.Errorf("%w: NewHandler()", er.ErrNilDependency)
	}
	return &Handler{service: service}, 	nil
}

func (s *Handler) Send(ctx context.Context, req *pb.TaskEvent) (*pb.Empty, error) {
	// 1. Маппинг типов
	var eventType en.NotificationType
      switch req.EventType {
      case pb.EventType_EVENT_TYPE_CREATED:
          eventType = en.TaskCreated
      case pb.EventType_EVENT_TYPE_DELETED:
          eventType = en.TaskDeleted
      default:
          return nil, fmt.Errorf("unknown event type: %v", req.EventType)
      }

      // 2. Вызов сервиса
      err := s.service.SendNotify(ctx, eventType, req.Id, req.Title)
      if err != nil {
          return nil, err
      }

      // 3. Возврат пустого ответа
      return &pb.Empty{}, nil
}