package usecases

import (
	"context"
	"fmt"
	"time"
	"encoding/json"

	"github.com/100bench/infr_training/internal/entities"
)

type TodoService struct {
	storage TaskStorage
	cache Cache
}

func NewTodoService(storage TaskStorage, cache Cache) (*TodoService, error) {
	if storage == nil {
		return nil, fmt.Errorf("NewTodoService: %w", entities.ErrNilDependency)
	}
	return &TodoService{
		storage: storage,
		cache: cache,
	}, nil
}

func (s *TodoService) CreateTask(ctx context.Context, title, description string) (*entities.Task, error) {
	task := entities.NewTask(title, description)

	err:= s.storage.Save(ctx, task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TodoService) GetTasks(ctx context.Context, id string) (*entities.Task, error) {
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
	err := s.storage.Delete(ctx, id)
	if err != nil {
		return err
	}
	// Invalidate cache after successful delete
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

