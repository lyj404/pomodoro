package app

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"

	"github.com/lyj404/pomodoro/internal/model"
	"github.com/lyj404/pomodoro/internal/service"
	"github.com/lyj404/pomodoro/internal/storage"
	"github.com/lyj404/pomodoro/internal/timer"
	"github.com/lyj404/pomodoro/internal/ui"
)

func Run() error {
	dataDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	appDataDir := filepath.Join(dataDir, "pomodoro")
	if err := os.MkdirAll(appDataDir, 0o755); err != nil {
		return err
	}

	db, err := storage.OpenDB(appDataDir)
	if err != nil {
		return err
	}
	defer db.Close()

	settingsRepo := storage.NewSettingsRepository(db)
	sessionRepo := storage.NewSessionRepository(db)

	settings, err := settingsRepo.Load()
	if err != nil {
		return err
	}

	fyneApp := fyneapp.NewWithID("pomodoro.desktop")
	fyneApp.Settings().SetTheme(ui.NewFocusTheme())
	fyneApp.SetIcon(ui.AppIcon())
	win := fyneApp.NewWindow("Pomodoro")
	win.SetIcon(ui.AppIcon())
	win.Resize(fyne.NewSize(420, 760))

	timerManager := timer.NewManager(settings)
	pomodoroService := service.NewPomodoroService(timerManager, settingsRepo, sessionRepo)
	statsService := service.NewStatsService(sessionRepo)
	currentStats, err := statsService.TodayStats()
	if err != nil {
		return err
	}

	var view *ui.MainView
	renderSnapshot := func() {
		view.Render(timerManager.Snapshot(), currentStats)
	}
	reloadStatsAndRender := func() {
		stats, statsErr := statsService.TodayStats()
		if statsErr != nil {
			view.ShowError(statsErr)
			return
		}
		currentStats = stats
		renderSnapshot()
	}
	view = ui.NewMainView(win, ui.MainCallbacks{
		OnStart: pomodoroService.Start,
		OnPause: pomodoroService.Pause,
		OnReset: func() {
			if err := pomodoroService.Reset(); err != nil {
				view.ShowError(err)
				return
			}
			reloadStatsAndRender()
		},
		OnSkip: func() {
			if err := pomodoroService.Skip(); err != nil {
				view.ShowError(err)
				return
			}
			reloadStatsAndRender()
		},
		OnOpenSettings: func() {
			current, err := settingsRepo.Load()
			if err != nil {
				view.ShowError(err)
				return
			}
			ui.ShowSettingsDialog(win, current, func(nextSettings model.Settings) {
				if err := pomodoroService.UpdateSettings(nextSettings); err != nil {
					view.ShowError(err)
					return
				}
				renderSnapshot()
			})
		},
		OnOpenHistory: func() {
			loadSessions := func() ([]model.Session, error) {
				return statsService.RecentSessions(50)
			}

			sessions, err := loadSessions()
			if err != nil {
				view.ShowError(err)
				return
			}
			ui.ShowHistoryDialog(win, sessions, statsService.DeleteSessions, loadSessions, func() {
				reloadStatsAndRender()
			})
		},
	})

	renderSnapshot()

	go func() {
		for event := range timerManager.Events() {
			if err := pomodoroService.HandleTimerEvent(event); err != nil {
				fyne.Do(func() {
					view.ShowError(fmt.Errorf("保存会话失败: %w", err))
				})
			}

			fyne.Do(func() {
				if event.Type == timer.EventPhaseFinished {
					title := "阶段完成"
					content := fmt.Sprintf("%s 已结束", ui.LocalModeLabel(event.Snapshot.Mode))
					fyneApp.SendNotification(fyne.NewNotification(title, content))
					view.ShowPhaseFinished(event.Snapshot)
					reloadStatsAndRender()
					return
				}
				renderSnapshot()
			})
		}
	}()

	win.ShowAndRun()
	return nil
}
