package tasker

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID        uuid.UUID `json:"id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"timestamp"`
}

func (t *Task) setRunning() {
	t.Status = "running"
}

func (t *Task) setFinished() {
	t.Status = "finished"
}

func (t *Task) finished() bool {
	return t.Status == "finished"
}
