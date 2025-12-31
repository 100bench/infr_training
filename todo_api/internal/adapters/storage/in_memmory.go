package storage

import (
	"context"
	"encoding/binary"
	"bytes"

	"github.com/100bench/infr_training/internal/entities"
)

type InMemory struct {
	data map[string][]byte
}

func NewInMemory() *InMemory {
	return &InMemory{
		data: make(map[string][]byte),
	}
}

func (s *InMemory) Save(ctx context.Context, task *entities.Task) error {
	buf:= new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, task.ID)
	binary.Write(buf, binary.LittleEndian, task.Title)
	binary.Write(buf, binary.LittleEndian, task.Description)
	binary.Write(buf, binary.LittleEndian, task.Completed)
	s.data[task.ID] = buf.Bytes()
	return nil
}

func (s *InMemory) Load(ctx context.Context, id string) (*entities.Task, error) {
	data, exists := s.data[id]
	if !exists {
		return nil, nil
	}
	buf := bytes.NewReader(data)
	var task entities.Task
	binary.Read(buf, binary.LittleEndian, &task.ID)
	binary.Read(buf, binary.LittleEndian, &task.Title)
	binary.Read(buf, binary.LittleEndian, &task.Description)
	binary.Read(buf, binary.LittleEndian, &task.Completed)
	return &task, nil
}

func (s *InMemory) Delete(ctx context.Context, id string) error {
	delete(s.data, id)
	return nil
}