package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

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
	ui.ApplyTheme(settings.Theme)
	ui.SetLang(settings.Language)
	if settings.Theme == "light" {
		fyneApp.Settings().SetTheme(ui.NewLightTheme())
	} else {
		fyneApp.Settings().SetTheme(ui.NewFocusTheme())
	}
	fyneApp.SetIcon(ui.AppIcon())
	win := fyneApp.NewWindow("Pomodoro")
	win.SetIcon(ui.AppIcon())
	win.Resize(fyne.NewSize(420, 760))
	win.SetFixedSize(true)

	timerManager := timer.NewManager(settings)
	pomodoroService := service.NewPomodoroService(timerManager, settingsRepo, sessionRepo)
	statsService := service.NewStatsService(sessionRepo)

	var currentSettings model.Settings
	var settingsMu sync.Mutex
	var settingsSaveTimer *time.Timer

	persistSettings := func(s model.Settings) {
		go func() {
			if err := settingsRepo.Save(s); err != nil {
				fmt.Printf("failed to persist settings: %v\n", err)
			}
		}()
	}

	saveSettingsDebounced := func(s model.Settings) {
		settingsMu.Lock()
		currentSettings = s
		if settingsSaveTimer != nil {
			settingsSaveTimer.Stop()
		}
		settingsSaveTimer = time.AfterFunc(500*time.Millisecond, func() {
			settingsMu.Lock()
			persistSettings(currentSettings)
			settingsMu.Unlock()
		})
		settingsMu.Unlock()
	}

	currentSettings = settings

	var currentStats storage.TodayStats
	var statsLoaded bool

	var view *ui.MainView
	renderSnapshot := func() {
		if !statsLoaded {
			stats, err := statsService.TodayStats()
			if err == nil {
				currentStats = stats
				statsLoaded = true
			}
		}
		view.Render(timerManager.Snapshot(), currentStats)
	}
	reloadStatsAndRender := func() {
		stats, statsErr := statsService.TodayStats()
		if statsErr != nil {
			view.ShowError(statsErr)
			return
		}
		currentStats = stats
		statsLoaded = true
		renderSnapshot()
	}
	invalidateStatsAndRender := func() {
		statsLoaded = false
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
			renderSnapshot()
		case fyne.KeyS:
			if err := pomodoroService.Skip(); err != nil {
				view.ShowError(err)
				return
			}
			renderSnapshot()
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
					renderSnapshot()
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
			renderSnapshot()
		},
		OnSkip: func() {
			if err := pomodoroService.Skip(); err != nil {
				view.ShowError(err)
				return
			}
			renderSnapshot()
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
				ui.ApplyTheme(nextSettings.Theme)
				ui.SetLang(nextSettings.Language)
				if nextSettings.Theme == "light" {
					fyneApp.Settings().SetTheme(ui.NewLightTheme())
				} else {
					fyneApp.Settings().SetTheme(ui.NewFocusTheme())
				}
				view.RefreshColors()
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
		OnToggleTheme: func() {
			settingsMu.Lock()
			newTheme := "dark"
			if currentSettings.Theme == "dark" {
				newTheme = "light"
			}
			currentSettings.Theme = newTheme
			saveSettingsDebounced(currentSettings)
			settingsMu.Unlock()

			ui.ApplyTheme(newTheme)
			if newTheme == "light" {
				fyneApp.Settings().SetTheme(ui.NewLightTheme())
			} else {
				fyneApp.Settings().SetTheme(ui.NewFocusTheme())
			}
			view.RefreshColors()
			renderSnapshot()
		},
		OnToggleLang: func(lang string) {
			settingsMu.Lock()
			currentSettings.Language = lang
			saveSettingsDebounced(currentSettings)
			settingsMu.Unlock()

			ui.SetLang(lang)
			view.RefreshText()
			renderSnapshot()
		},
	})

	renderSnapshot()

	settingsCopy := settings
	uiRefreshTicker := time.NewTicker(time.Second)
	defer uiRefreshTicker.Stop()

	go func() {
		for {
			select {
			case <-uiRefreshTicker.C:
				fyne.Do(func() {
					renderSnapshot()
				})
			case event, ok := <-timerManager.Events():
				if !ok {
					return
				}
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
						invalidateStatsAndRender()
					}
				})
			}
		}
	}()

	win.ShowAndRun()
	return nil
}
