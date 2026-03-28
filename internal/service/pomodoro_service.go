package service

import (
	"time"

	"github.com/lyj404/pomodoro/internal/model"
	"github.com/lyj404/pomodoro/internal/storage"
	"github.com/lyj404/pomodoro/internal/timer"
)

type PomodoroService struct {
	timer        *timer.Manager
	settingsRepo *storage.SettingsRepository
	sessionRepo  *storage.SessionRepository

	currentPhaseStart time.Time
}

func NewPomodoroService(
	timerManager *timer.Manager,
	settingsRepo *storage.SettingsRepository,
	sessionRepo *storage.SessionRepository,
) *PomodoroService {
	return &PomodoroService{
		timer:        timerManager,
		settingsRepo: settingsRepo,
		sessionRepo:  sessionRepo,
	}
}

func (s *PomodoroService) Start() {
	snap := s.timer.Snapshot()
	if snap.Status == timer.StatusIdle {
		s.currentPhaseStart = time.Now()
	}
	s.timer.Start()
}

func (s *PomodoroService) Pause() {
	s.timer.Pause()
}

func (s *PomodoroService) Reset() error {
	if err := s.persistCurrentSession(false); err != nil {
		return err
	}
	s.timer.Reset()
	return nil
}

func (s *PomodoroService) Skip() error {
	if err := s.persistCurrentSession(false); err != nil {
		return err
	}
	s.timer.Skip()
	s.currentPhaseStart = time.Time{}
	return nil
}

func (s *PomodoroService) UpdateSettings(settings model.Settings) error {
	if err := s.settingsRepo.Save(settings); err != nil {
		return err
	}
	s.timer.UpdateSettings(settings)
	return nil
}

func (s *PomodoroService) HandleTimerEvent(event timer.Event) error {
	if event.Type != timer.EventPhaseFinished {
		return nil
	}

	endedAt := time.Now()
	startedAt := s.currentPhaseStart
	if startedAt.IsZero() {
		startedAt = endedAt.Add(-time.Duration(event.Snapshot.TotalSeconds) * time.Second)
	}

	session := model.Session{
		Mode:           event.Snapshot.Mode,
		PlannedSeconds: event.Snapshot.TotalSeconds,
		ActualSeconds:  event.Snapshot.TotalSeconds,
		StartedAt:      startedAt,
		EndedAt:        &endedAt,
		Completed:      true,
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return err
	}

	s.currentPhaseStart = endedAt
	return nil
}

func (s *PomodoroService) persistCurrentSession(completed bool) error {
	snap := s.timer.Snapshot()
	if s.currentPhaseStart.IsZero() || snap.Status == timer.StatusIdle {
		return nil
	}

	endedAt := time.Now()
	actualSeconds := snap.TotalSeconds - snap.RemainingSeconds
	if actualSeconds <= 0 {
		return nil
	}

	session := model.Session{
		Mode:           snap.Mode,
		PlannedSeconds: snap.TotalSeconds,
		ActualSeconds:  actualSeconds,
		StartedAt:      s.currentPhaseStart,
		EndedAt:        &endedAt,
		Completed:      completed,
	}

	s.currentPhaseStart = time.Time{}
	return s.sessionRepo.Create(session)
}
