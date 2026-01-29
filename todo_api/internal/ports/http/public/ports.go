package public

import (
	"context"
	"github.com/100bench/infr_training/todo_api/internal/entities"
)

type ServicePort interface {
	CreateTask(ctx context.Context, title, description string) (*entities.Task, error)
	GetTask(ctx context.Context, id string) (*entities.Task, error)
	DeleteTask(ctx context.Context, id string) error
}
