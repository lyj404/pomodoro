package timer

import (
	"sync"
	"time"

	"github.com/lyj404/pomodoro/internal/model"
)

type EventType string

const (
	EventTick          EventType = "tick"
	EventStateChanged  EventType = "state_changed"
	EventPhaseFinished EventType = "phase_finished"
)

type Event struct {
	Type     EventType
	Snapshot Snapshot
}

type Manager struct {
	mu                 sync.Mutex
	settings           model.Settings
	mode               model.SessionMode
	status             Status
	totalSeconds       int
	remainingSeconds   int
	completedPomodoros int
	ticker             *time.Ticker
	stopCh             chan struct{}
	tickerWg           sync.WaitGroup
	events             chan Event
}

func NewManager(settings model.Settings) *Manager {
	m := &Manager{
		settings: settings,
		mode:     model.SessionModeWork,
		status:   StatusIdle,
		events:   make(chan Event, 32),
	}
	m.applyModeLocked(model.SessionModeWork)
	return m
}

func (m *Manager) Events() <-chan Event {
	return m.events
}

func (m *Manager) Snapshot() Snapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.snapshotLocked()
}

func (m *Manager) UpdateSettings(settings model.Settings) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.settings = settings
	if m.status == StatusIdle {
		m.applyModeLocked(m.mode)
		m.emitLocked(EventStateChanged)
	}
}

func (m *Manager) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status == StatusRunning {
		return
	}
	if m.status == StatusIdle {
		m.applyModeLocked(m.mode)
	}

	m.status = StatusRunning
	m.startTickerLocked()
	m.emitLocked(EventStateChanged)
}

func (m *Manager) Pause() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status != StatusRunning {
		return
	}

	m.stopTickerLocked()
	m.status = StatusPaused
	m.emitLocked(EventStateChanged)
}

func (m *Manager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stopTickerLocked()
	m.status = StatusIdle
	m.applyModeLocked(m.mode)
	m.emitLocked(EventStateChanged)
}

func (m *Manager) Skip() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stopTickerLocked()
	nextMode := m.nextModeLocked(true)
	m.mode = nextMode
	m.status = StatusIdle
	m.applyModeLocked(nextMode)
	m.emitLocked(EventStateChanged)
}

func (m *Manager) stopTickerLocked() {
	if m.ticker != nil {
		m.ticker.Stop()
		m.ticker = nil
	}
	if m.stopCh != nil {
		close(m.stopCh)
		m.stopCh = nil
	}
	m.tickerWg.Wait()
}

func (m *Manager) startTickerLocked() {
	m.stopTickerLocked()
	m.ticker = time.NewTicker(time.Second)
	m.stopCh = make(chan struct{})

	m.tickerWg.Add(1)
	go func(t *time.Ticker, stopCh chan struct{}) {
		defer m.tickerWg.Done()
		for {
			select {
			case <-t.C:
				m.onTick()
			case <-stopCh:
				return
			}
		}
	}(m.ticker, m.stopCh)
}

func (m *Manager) onTick() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status != StatusRunning {
		return
	}

	if m.remainingSeconds > 0 {
		m.remainingSeconds--
		m.emitLocked(EventTick)
	}

	if m.remainingSeconds > 0 {
		return
	}

	finishedSnapshot := m.snapshotLocked()

	if m.mode == model.SessionModeWork {
		m.completedPomodoros++
	}

	nextMode := m.nextModeLocked(false)
	m.stopTickerLocked()
	m.mode = nextMode
	m.status = StatusIdle
	m.applyModeLocked(nextMode)

	m.events <- Event{
		Type:     EventPhaseFinished,
		Snapshot: finishedSnapshot,
	}
	m.emitLocked(EventStateChanged)

	if m.settings.AutoStartNextPhase {
		m.status = StatusRunning
		m.startTickerLocked()
		m.emitLocked(EventStateChanged)
	}
}

func (m *Manager) nextModeLocked(skipped bool) model.SessionMode {
	switch m.mode {
	case model.SessionModeWork:
		if skipped {
			return model.SessionModeShortBreak
		}
		if m.settings.LongBreakInterval > 0 && (m.completedPomodoros+1)%m.settings.LongBreakInterval == 0 {
			return model.SessionModeLongBreak
		}
		return model.SessionModeShortBreak
	case model.SessionModeShortBreak, model.SessionModeLongBreak:
		return model.SessionModeWork
	default:
		return model.SessionModeWork
	}
}

func (m *Manager) applyModeLocked(mode model.SessionMode) {
	m.mode = mode
	m.totalSeconds = m.durationForModeLocked(mode)
	m.remainingSeconds = m.totalSeconds
}

func (m *Manager) durationForModeLocked(mode model.SessionMode) int {
	switch mode {
	case model.SessionModeShortBreak:
		return m.settings.ShortBreakMinutes * 60
	case model.SessionModeLongBreak:
		return m.settings.LongBreakMinutes * 60
	default:
		return m.settings.WorkMinutes * 60
	}
}

func (m *Manager) emitLocked(eventType EventType) {
	select {
	case m.events <- Event{Type: eventType, Snapshot: m.snapshotLocked()}:
	default:
	}
}

func (m *Manager) snapshotLocked() Snapshot {
	return Snapshot{
		Mode:               m.mode,
		Status:             m.status,
		RemainingSeconds:   m.remainingSeconds,
		TotalSeconds:       m.totalSeconds,
		CompletedPomodoros: m.completedPomodoros,
	}
}
