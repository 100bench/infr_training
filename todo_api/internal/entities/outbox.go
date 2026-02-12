package entities

import (
	"time"
	"github.com/google/uuid"
)


const (
	EventTypeTaskCreated = "task.created"
	EventTypeTaskDeleted = "task.deleted"
)

type OutboxEvent struct {
	ID uuid.UUID `json:"id"`
	TaskID uuid.UUID `json:"task_id"`
	EventType string `json:"event_type"`
	Payload []byte `json:"payload"`
	Sent bool `json:"sent"`
	CreatedAt time.Time `json:"created_at"`
}