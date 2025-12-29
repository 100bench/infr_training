package public

import (
	"context"
	"github.com/100bench/infr_training/internal/entities"
)

type ServicePort interface {
	CreateTask(ctx context.Context, title, description string) error
	GetTasks(ctx context.Context, id string) (*entities.Task, error)
	DeleteTask(ctx context.Context, id string) error
}
