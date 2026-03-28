package timer

import "github.com/lyj404/pomodoro/internal/model"

type Status string

const (
	StatusIdle    Status = "idle"
	StatusRunning Status = "running"
	StatusPaused  Status = "paused"
)

type Snapshot struct {
	Mode               model.SessionMode
	Status             Status
	RemainingSeconds   int
	TotalSeconds       int
	CompletedPomodoros int
}
