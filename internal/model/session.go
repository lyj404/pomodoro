package model

import "time"

type SessionMode string

const (
	SessionModeWork       SessionMode = "work"
	SessionModeShortBreak SessionMode = "short_break"
	SessionModeLongBreak  SessionMode = "long_break"
)

type Session struct {
	ID             int64
	Mode           SessionMode
	PlannedSeconds int
	ActualSeconds  int
	StartedAt      time.Time
	EndedAt        *time.Time
	Completed      bool
}
