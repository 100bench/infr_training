package entities

type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool `json:"completed"`
}

func NewTask(id, title, description string) *Task {
	return &Task{
		ID:          id,
		Title:       title,
		Description: description,
		Completed:   false,
	}
}