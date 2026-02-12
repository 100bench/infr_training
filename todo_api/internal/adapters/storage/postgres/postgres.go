package postgres

import (
	"context"
	"fmt"
	"errors"	
	"encoding/json"

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

// func (t *TaskStorage) Save(ctx context.Context, task *en.Task) error {
// 	const q = `INSERT INTO tasks (title, description, completed)
// 	VALUES ($1, $2, $3) 
// 	RETURNING id`
// 	err := t.pool.QueryRow(ctx, q, task.Title, task.Description, task.Completed).Scan(&task.ID)
// 	if err != nil {
// 		return fmt.Errorf("%w: Taskstorage.Save()", err)
// 	}
// 	return nil
// }

// version with outbox pattern
func (t *TaskStorage) Save(ctx context.Context, task *en.Task) error {
	tx, err:= t.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)


	const q = `INSERT INTO tasks (title, description, completed)
	VALUES ($1, $2, $3) 
	RETURNING id`
	err = tx.QueryRow(ctx, q, task.Title, task.Description, task.Completed).Scan(&task.ID)
	if err != nil {
		return fmt.Errorf("Taskstorage.Save(): %w", err)
	}


	type TaskEventPayload struct {
    	ID    string `json:"id"`
    	Title string `json:"title"`
	}

	payload, _ := json.Marshal(TaskEventPayload{
		ID:    task.ID.String(),
		Title: task.Title,
	})
	const outboxQ = `INSERT INTO outbox (task_id, event_type, payload) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, outboxQ, task.ID, en.EventTypeTaskCreated, payload)
	if err != nil {
		return fmt.Errorf("failed to insert into outbox: %w", err)
	}


	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
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

// func (t *TaskStorage) Delete(ctx context.Context, id string) (string, error) {
// 	const q = `
// 		DELETE FROM tasks
// 		WHERE id = $1
// 		RETURNING title
// 	`

// 	var title string
// 	err := t.pool.QueryRow(ctx, q, id).Scan(&title)
// 	if err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			return "", er.ErrEntityNotFound
// 		}
// 		return "", fmt.Errorf("TaskStorage.Delete: %w", err)
// 	}

// 	return title, nil
// }


// version with outbox pattern
func (t *TaskStorage) Delete(ctx context.Context, id string) (string, error) {
	tx, err:= t.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	
	const q = `
		DELETE FROM tasks
		WHERE id = $1
		RETURNING title
	`

	var title string
	err = tx.QueryRow(ctx, q, id).Scan(&title)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", er.ErrEntityNotFound
		}
		return "", fmt.Errorf("TaskStorage.Delete: %w", err)
	}

	type TaskEventPayload struct {
    	ID    string `json:"id"`
    	Title string `json:"title"`
	}

	payload, _ := json.Marshal(TaskEventPayload{
		ID:    id,
		Title: title,
	})
	const outboxQ = `INSERT INTO outbox (event_type, payload) VALUES ($1, $2)`
	_, err = tx.Exec(ctx, outboxQ, en.EventTypeTaskDeleted, payload)
	if err != nil {
		return "", fmt.Errorf("failed to insert into outbox: %w", err)
	}


	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}
	return title, nil
}