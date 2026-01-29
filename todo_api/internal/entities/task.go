package entities

import "github.com/google/uuid"

type Task struct {
	ID          uuid.UUID `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool `json:"completed"`
}

func NewTask(title, description string) *Task {
	return &Task{
		Title:       title,
		Description: description,
		Completed:   false,
	}
}