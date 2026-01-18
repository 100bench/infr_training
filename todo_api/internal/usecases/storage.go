package usecases


import ("context"
	"github.com/100bench/infr_training/todo_api/internal/entities"
)

type TaskStorage interface {
	Save(ctx context.Context, task *entities.Task) error
	Load(ctx context.Context, id string) (*entities.Task, error)
	Delete(ctx context.Context, id string) error
}