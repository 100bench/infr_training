package postgres

import (
	"context"
	"fmt"
	"errors"

	en "github.com/100bench/infr_training/todo_api/internal/entities"
	er "github.com/100bench/infr_training/pkg/errors"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type TaskStorage struct {
	pool *pgxpool.Pool
}

func NewTaskStorage(ctx context.Context, dsn string) (*TaskStorage, error) {
	pool, err:= pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	return &TaskStorage{
		pool: pool,
	}, nil
}

func (t *TaskStorage) Close() {
	t.pool.Close()
}

func (t *TaskStorage) Save(ctx context.Context, task *en.Task) error {
	const q = `INSERT INTO tasks (title, description, completed) 
	VALUES ($1, $2, $3) 
	RETURNING id`
	err := t.pool.QueryRow(ctx, q, task.Title, task.Description, task.Completed).Scan(&task.ID)
	if err != nil {
		return fmt.Errorf("%w: Taskstorage.Save()", err)
	}
	return nil
}

func (t *TaskStorage) Load(ctx context.Context, id string) (*en.Task, error) {
	const q = `SELECT id, title, description, completed FROM tasks WHERE id=$1`
	row := t.pool.QueryRow(ctx, q, id)

	task := &en.Task{}
	err := row.Scan(&task.ID, &task.Title, &task.Description, &task.Completed)
	if err == pgx.ErrNoRows{
		return nil, fmt.Errorf("%w: Storage.Load()", er.ErrEntityNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: Taskstorage.Load()", err)
	}
	return task, nil
}

func (t *TaskStorage) Delete(ctx context.Context, id string) (string, error) {
	const q = `
		DELETE FROM tasks
		WHERE id = $1
		RETURNING title
	`

	var title string
	err := t.pool.QueryRow(ctx, q, id).Scan(&title)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", er.ErrEntityNotFound
		}
		return "", fmt.Errorf("TaskStorage.Delete: %w", err)
	}

	return title, nil
}
