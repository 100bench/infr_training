package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	er "github.com/100bench/infr_training/pkg/errors"
	"github.com/100bench/infr_training/pkg/logger"
	pb "github.com/100bench/infr_training/pkg/proto/gen/notification/v1"
	"github.com/100bench/infr_training/todo_api/internal/entities"
)

type TodoService struct {
	storage TaskStorage
	cache Cache
	notifier Notifier
	log logger.Logger
}

func NewTodoService(storage TaskStorage, cache Cache, notifier Notifier, logger logger.Logger) (*TodoService, error) {
	if storage == nil {
		return nil, fmt.Errorf("NewTodoService storage: %w", er.ErrNilDependency)
	}
	if cache == nil {
		return nil, fmt.Errorf("NewTodoService cache: %w", er.ErrNilDependency)
	}
	if notifier == nil {
		return nil, fmt.Errorf("NewTodoService notifier: %w", er.ErrNilDependency)
	}
	if logger == nil {
		return nil, fmt.Errorf("NewTodoService logger: %w", er.ErrNilDependency)
	}
	return &TodoService{
		storage: storage,
		cache: cache,
		notifier: notifier,
		log: logger,
	}, nil
}

func (s *TodoService) CreateTask(ctx context.Context, title, description string) (*entities.Task, error) {
	task := entities.NewTask(title, description)

	err:= s.storage.Save(ctx, task)
	if err != nil {
		return nil, err
	}

	err = s.notifier.Send(ctx, task.ID.String(), title, pb.EventType_EVENT_TYPE_CREATED)
	if err != nil{
		s.log.Warn("failed to send create notification")
	}

	return task, nil
}

func (s *TodoService) GetTask(ctx context.Context, id string) (*entities.Task, error) {
	key := fmt.Sprintf("task:%s", id)

	// Try cache first
	if s.cache != nil {
		cached, err := s.cache.Get(ctx, key)
		if err == nil {
			return decodeTask(cached)
		}
	}

	// Fallback to storage
	task, err := s.storage.Load(ctx, id)
	if err != nil {
		return nil, err
	}

	// Save to cache
	if s.cache != nil {
		if encodedTask, err := encodeTask(task); err == nil {
			s.cache.Set(ctx, key, encodedTask, time.Minute*5)
		}
	}
	return task, nil
}

func (s *TodoService) DeleteTask(ctx context.Context, id string) error {
	title, err := s.storage.Delete(ctx, id)
	if err != nil {
		return err
	}

	err = s.notifier.Send(ctx, id, title, pb.EventType_EVENT_TYPE_DELETED)
	if err != nil{
		s.log.Warn("failed to send delete notification")
	}
	
	if s.cache != nil {
		key := fmt.Sprintf("task:%s", id)
		s.cache.Delete(ctx, key)
	}
	return nil
}

func encodeTask(t *entities.Task) (string, error) {
    b, err := json.Marshal(t)
    if err != nil {
        return "", err
    }
    return string(b), nil
}

func decodeTask(s string) (*entities.Task, error) {
    var t entities.Task
    err := json.Unmarshal([]byte(s), &t)
    return &t, err
}

