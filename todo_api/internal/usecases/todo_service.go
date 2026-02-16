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
	"golang.org/x/sync/singleflight"
)

type TodoService struct{
	ctx context.Context
	storage TaskStorage
	cache Cache
	notifier Notifier
	log logger.Logger
	group singleflight.Group
}

func NewTodoService(ctx context.Context, storage TaskStorage, cache Cache, notifier Notifier, logger logger.Logger) (*TodoService, error) {
	if ctx == nil{
		return nil, fmt.Errorf("NewTodoService ctx: %w", er.ErrNilDependency)
	}
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
		ctx: ctx,
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

	v, err, shared := s.group.Do(key, func() (interface{}, error) {
		val, err:= s.cache.Get(s.ctx, key)
		if err == nil {
			return decodeTask(val)
		}

		task, err := s.storage.Load(s.ctx, id)
		if err != nil {
			return nil, fmt.Errorf("storage load: %w", err)
		}

		if encodedTask, err := encodeTask(task); err == nil {
			s.cache.Set(s.ctx, key, encodedTask, time.Minute*5)
		}
		return task, nil
	})

	if shared { 
		s.log.Info("singleflight: request deduplicated for key", logger.Field{Key: "key", Value: key})
		}

	if err != nil {
		return nil, err
	}
	return v.(*entities.Task), nil
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

