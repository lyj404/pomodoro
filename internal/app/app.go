package app

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/gen2brain/beeep"

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

	fyneApp := app.NewWithID("pomodoro.desktop")
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

	win.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) {
		switch e.Name {
		case fyne.KeySpace:
			snapshot := timerManager.Snapshot()
			if snapshot.Status == timer.StatusRunning {
				pomodoroService.Pause()
			} else {
				pomodoroService.Start()
			}
		case fyne.KeyR:
			if err := pomodoroService.Reset(); err != nil {
				view.ShowError(err)
				return
			}
			reloadStatsAndRender()
		case fyne.KeyS:
			if err := pomodoroService.Skip(); err != nil {
				view.ShowError(err)
				return
			}
			reloadStatsAndRender()
		}
	})

	win.SetCloseIntercept(func() {
		win.Hide()
	})

	trayMenu := &fyne.Menu{
		Items: []*fyne.MenuItem{
			{
				Label: "显示窗口",
				Action: func() {
					win.Show()
					win.RequestFocus()
				},
			},
			{IsSeparator: true},
			{
				Label: "开始/暂停",
				Action: func() {
					snapshot := timerManager.Snapshot()
					if snapshot.Status == timer.StatusRunning {
						pomodoroService.Pause()
					} else {
						pomodoroService.Start()
					}
				},
			},
			{
				Label: "重置",
				Action: func() {
					if err := pomodoroService.Reset(); err != nil {
						view.ShowError(err)
						return
					}
					reloadStatsAndRender()
				},
			},
			{IsSeparator: true},
			{
				Label: "退出",
				Action: func() {
					win.Close()
					fyneApp.Quit()
				},
			},
		},
	}
	var desktopApp desktop.App
	var ok bool
	if desktopApp, ok = fyneApp.(desktop.App); ok {
		desktopApp.SetSystemTrayMenu(trayMenu)
		desktopApp.SetSystemTrayIcon(ui.AppIcon())
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
			total, err := statsService.CountTotalSessions()
			if err != nil {
				view.ShowError(err)
				return
			}

			loadSessions := func(offset int) ([]model.Session, int, error) {
				sessions, total, err := statsService.RecentSessionsPaginated(offset, 20)
				return sessions, total, err
			}

			sessions, _, err := loadSessions(0)
			if err != nil {
				view.ShowError(err)
				return
			}
			ui.ShowHistoryDialog(win, sessions, total, statsService.DeleteSessions, loadSessions, func() {
				reloadStatsAndRender()
				total, _ = statsService.CountTotalSessions()
			})
		},
	})

	renderSnapshot()

	settingsCopy := settings
	go func() {
		for event := range timerManager.Events() {
			if err := pomodoroService.HandleTimerEvent(event); err != nil {
				fyne.Do(func() {
					view.ShowError(fmt.Errorf("保存会话失败: %w", err))
				})
			}

			fyne.Do(func() {
				if event.Type == timer.EventPhaseFinished {
					if settingsCopy.SoundEnabled {
						beeep.Notify("Pomodoro", "阶段完成", "default")
					}
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
