package usecases

import (
	"context"
	"github.com/100bench/infr_training/internal/entities"
	"github.com/google/uuid"
)

type TodoService struct {
	storage Storage
}

func NewTodoService(storage Storage) *TodoService {
	return &TodoService{
		storage: storage,
	}
}

func (s *TodoService) CreateTask(ctx context.Context, title, description string) error {
	task := &entities.Task{
		Title:       title,
		Description: description,
		Completed:   false,
	}
	task.ID = uuid.New().String()

	err:= s.storage.Save(ctx, task)
	if err != nil {
		return err
	}
	return nil
}

func (s *TodoService) GetTasks(ctx context.Context, id string) (*entities.Task, error) {
	task, err := s.storage.Load(ctx, id)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TodoService) DeleteTask(ctx context.Context, id string) error {
	err := s.storage.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}
