package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/lyj404/pomodoro/internal/model"
	"github.com/lyj404/pomodoro/internal/storage"
	"github.com/lyj404/pomodoro/internal/timer"
)

type MainView struct {
	window             fyne.Window
	background         *canvas.Rectangle
	accentBar          *canvas.Rectangle
	timerCard          *canvas.Rectangle
	modeLabel          *canvas.Text
	timeLabel          *canvas.Text
	statusLabel        *canvas.Text
	phaseHintLabel     *canvas.Text
	focusCountValue    *canvas.Text
	focusDurationValue *canvas.Text
	streakValue        *canvas.Text
	startBtn           *ActionTile
	pauseBtn           *ActionTile
	resetBtn           *ActionTile
	skipBtn            *ActionTile
	settingsBtn        *widget.Button
	historyBtn         *widget.Button
	rootContent        *fyne.Container
	scroll             *container.Scroll
	layoutTier         layoutTier
	firstRender        bool
	prevSnapshot       timer.Snapshot
	prevTotalSeconds   int
	callbacks          MainCallbacks
}

type layoutTier int

const (
	layoutTierUnknown layoutTier = iota
	layoutTierTiny
	layoutTierCompact
	layoutTierMedium
	layoutTierLarge
)

type MainCallbacks struct {
	OnStart        func()
	OnPause        func()
	OnReset        func()
	OnSkip         func()
	OnOpenSettings func()
	OnOpenHistory  func()
}

func NewMainView(win fyne.Window, callbacks MainCallbacks) *MainView {
	view := &MainView{
		window:             win,
		background:         canvas.NewRectangle(appBackgroundColor),
		accentBar:          canvas.NewRectangle(workAccentColor),
		timerCard:          canvas.NewRectangle(cardBackgroundColor),
		modeLabel:          canvas.NewText("工作中", workAccentColor),
		timeLabel:          canvas.NewText("25:00", nordText),
		statusLabel:        canvas.NewText("待开始", nordText),
		phaseHintLabel:     canvas.NewText("准备进入一段专注时刻", nordText),
		focusCountValue:    canvas.NewText("0", nordText),
		focusDurationValue: canvas.NewText("00:00", nordText),
		streakValue:        canvas.NewText("0", nordText),
		layoutTier:         layoutTierUnknown,
		firstRender:        true,
		callbacks:          callbacks,
	}

	view.timerCard.CornerRadius = 22

	view.modeLabel.Alignment = fyne.TextAlignCenter
	view.modeLabel.TextSize = 16
	view.modeLabel.TextStyle = fyne.TextStyle{Bold: true}

	view.timeLabel.Alignment = fyne.TextAlignCenter
	view.timeLabel.TextSize = 48
	view.timeLabel.TextStyle = fyne.TextStyle{Bold: true}

	view.statusLabel.Alignment = fyne.TextAlignCenter
	view.statusLabel.TextSize = 14

	view.phaseHintLabel.Alignment = fyne.TextAlignCenter
	view.phaseHintLabel.TextSize = 12

	view.focusCountValue.TextSize = 24
	view.focusCountValue.TextStyle = fyne.TextStyle{Bold: true}
	view.focusCountValue.Alignment = fyne.TextAlignCenter
	view.focusDurationValue.TextSize = 24
	view.focusDurationValue.TextStyle = fyne.TextStyle{Bold: true}
	view.focusDurationValue.Alignment = fyne.TextAlignCenter
	view.streakValue.TextSize = 24
	view.streakValue.TextStyle = fyne.TextStyle{Bold: true}
	view.streakValue.Alignment = fyne.TextAlignCenter

	view.startBtn = NewActionTile("开始专注", theme.MediaPlayIcon(), primaryButtonColor, primaryButtonTextColor, callbacks.OnStart)
	view.pauseBtn = NewActionTile("暂停", theme.MediaPauseIcon(), pauseButtonColor, pauseButtonTextColor, callbacks.OnPause)
	view.resetBtn = NewActionTile("重置", theme.ViewRefreshIcon(), secondaryButtonColor, secondaryTextColor, callbacks.OnReset)
	view.skipBtn = NewActionTile("跳过", theme.MediaSkipNextIcon(), secondaryButtonColor, secondaryTextColor, callbacks.OnSkip)
	view.settingsBtn = widget.NewButtonWithIcon("", theme.SettingsIcon(), callbacks.OnOpenSettings)
	view.settingsBtn.Importance = widget.LowImportance
	view.historyBtn = widget.NewButtonWithIcon("", theme.HistoryIcon(), callbacks.OnOpenHistory)
	view.historyBtn.Importance = widget.LowImportance

	view.rootContent = container.NewVBox()
	view.scroll = container.NewVScroll(container.NewPadded(view.rootContent))

	toolbar := container.NewHBox(
		layout.NewSpacer(),
		view.settingsBtn,
		view.historyBtn,
	)

	win.SetContent(container.NewStack(
		view.background,
		container.NewBorder(toolbar, nil, nil, nil, container.NewBorder(view.accentBar, nil, nil, nil, view.scroll)),
	))

	view.ensureLayout(win.Canvas().Size().Width)
	view.startBtn.Refresh()
	view.pauseBtn.Refresh()
	view.resetBtn.Refresh()
	view.skipBtn.Refresh()
	return view
}

func (v *MainView) Render(snapshot timer.Snapshot, stats storage.TodayStats) {
	accent := accentColorForMode(snapshot.Mode)

	if snapshot.Mode != v.prevSnapshot.Mode {
		v.accentBar.FillColor = accent
		v.modeLabel.Color = accent
		v.modeLabel.Text = localModeLabel(snapshot.Mode)
		v.background.FillColor = backgroundForMode(snapshot.Mode)
		v.background.Refresh()
		v.accentBar.Refresh()
		v.modeLabel.Refresh()
	}

	if snapshot.Status != v.prevSnapshot.Status || snapshot.Mode != v.prevSnapshot.Mode {
		v.statusLabel.Text = localStatusLabel(snapshot.Status)
		v.phaseHintLabel.Text = phaseHint(snapshot)
		v.statusLabel.Refresh()
		v.phaseHintLabel.Refresh()
	}

	if snapshot.RemainingSeconds != v.prevSnapshot.RemainingSeconds || snapshot.TotalSeconds != v.prevTotalSeconds {
		v.timeLabel.Text = formatClock(snapshot.RemainingSeconds)
		v.timeLabel.Refresh()
		v.prevTotalSeconds = snapshot.TotalSeconds
	}

	if stats.CompletedPomodoros != int(v.prevSnapshot.CompletedPomodoros) || v.firstRender {
		v.focusCountValue.Text = fmt.Sprintf("%d", stats.CompletedPomodoros)
		v.focusCountValue.Refresh()
	}

	if stats.FocusSeconds != 0 || v.firstRender {
		v.focusDurationValue.Text = formatDuration(stats.FocusSeconds)
		v.focusDurationValue.Refresh()
	}

	if snapshot.CompletedPomodoros != v.prevSnapshot.CompletedPomodoros || v.firstRender {
		v.streakValue.Text = fmt.Sprintf("%d", snapshot.CompletedPomodoros)
		v.streakValue.Refresh()
	}

	running := snapshot.Status == timer.StatusRunning
	if running != (v.prevSnapshot.Status == timer.StatusRunning) {
		if running {
			v.startBtn.SetDisabled(true)
			v.pauseBtn.SetDisabled(false)
		} else {
			v.startBtn.SetDisabled(false)
			v.pauseBtn.SetDisabled(true)
		}
		v.startBtn.Refresh()
		v.pauseBtn.Refresh()
	}

	v.ensureLayout(v.window.Canvas().Size().Width)
	v.prevSnapshot = snapshot
	v.firstRender = false
}

func (v *MainView) ensureLayout(width float32) {
	nextTier := layoutTierForWidth(width)
	if nextTier == v.layoutTier {
		return
	}
	v.layoutTier = nextTier

	timerInner := container.NewVBox(
		layout.NewSpacer(),
		v.modeLabel,
		verticalGap(1),
		v.timeLabel,
		verticalGap(1),
		v.statusLabel,
		verticalGap(1),
		v.phaseHintLabel,
		layout.NewSpacer(),
	)

	timerCard := container.NewStack(v.timerCard, container.NewPadded(timerInner))

	statsCards := []fyne.CanvasObject{
		statCard("今日番茄", v.focusCountValue, "已完成"),
		statCard("专注时长", v.focusDurationValue, "工作阶段累计"),
		statCard("当前连击", v.streakValue, "连续完成"),
	}

	var primaryActions fyne.CanvasObject
	var secondaryActions fyne.CanvasObject
	var stats fyne.CanvasObject

	switch nextTier {
	case layoutTierLarge:
		primaryActions = container.NewGridWithColumns(2, v.startBtn, v.pauseBtn)
		secondaryActions = actionGroup(v.resetBtn, v.skipBtn)
		stats = container.NewGridWithColumns(3, statsCards...)
	case layoutTierMedium:
		primaryActions = container.NewGridWithColumns(2, v.startBtn, v.pauseBtn)
		secondaryActions = actionGroup(v.resetBtn, v.skipBtn)
		stats = container.NewGridWithColumns(2,
			container.NewVBox(statsCards[0], verticalGap(4), statsCards[1]),
			statsCards[2],
		)
	case layoutTierCompact:
		primaryActions = container.NewGridWithColumns(2, v.startBtn, v.pauseBtn)
		secondaryActions = actionGroup(v.resetBtn, v.skipBtn)
		stats = container.NewGridWithColumns(2,
			statsCards[0],
			statsCards[1],
			statsCards[2],
			verticalGap(1),
		)
	default:
		primaryActions = container.NewVBox(v.startBtn, verticalGap(4), v.pauseBtn)
		secondaryActions = actionGroupVertical(v.resetBtn, v.skipBtn)
		stats = container.NewVBox(
			statsCards[0],
			verticalGap(4),
			statsCards[1],
			verticalGap(4),
			statsCards[2],
		)
	}

	v.rootContent.Objects = []fyne.CanvasObject{
		verticalGap(2),
		timerCard,
		verticalGap(4),
		primaryActions,
		verticalGap(2),
		secondaryActions,
		verticalGap(3),
		stats,
	}
	v.rootContent.Refresh()
}

func layoutTierForWidth(width float32) layoutTier {
	switch {
	case width >= 620:
		return layoutTierLarge
	case width >= 460:
		return layoutTierMedium
	case width >= 360:
		return layoutTierCompact
	default:
		return layoutTierTiny
	}
}

func (v *MainView) ShowPhaseFinished(snapshot timer.Snapshot) {
	header := dialogHeader("阶段完成", fmt.Sprintf("%s 已结束", LocalModeLabel(snapshot.Mode)), accentColorForMode(snapshot.Mode))
	bodyText := canvas.NewText("你可以稍作调整，然后继续下一阶段。", nordSubtext)
	bodyText.TextSize = 13

	content := container.NewPadded(container.NewVBox(
		header,
		verticalGap(8),
		bodyText,
	))

	finishedDialog := dialog.NewCustom("阶段完成", "好的", content, v.window)
	finishedDialog.Resize(fyne.NewSize(440, 220))
	finishedDialog.Show()
}

func (v *MainView) ShowError(err error) {
	dialog.ShowError(err, v.window)
}

func statCard(title string, value *canvas.Text, hint string) fyne.CanvasObject {
	card := canvas.NewRectangle(cardBackgroundColor)
	card.CornerRadius = 20

	titleText := canvas.NewText(title, nordSubtext)
	titleText.TextSize = 11
	titleText.Alignment = fyne.TextAlignCenter

	hintText := canvas.NewText(hint, nordSubtext)
	hintText.TextSize = 9
	hintText.Alignment = fyne.TextAlignCenter

	body := container.NewVBox(
		container.NewCenter(titleText),
		container.NewCenter(value),
		container.NewCenter(hintText),
	)

	return container.NewStack(card, container.NewPadded(body))
}

func actionGroup(buttons ...fyne.CanvasObject) fyne.CanvasObject {
	card := canvas.NewRectangle(toolbarCardColor)
	card.CornerRadius = 20
	return container.NewStack(
		card,
		container.NewPadded(container.NewGridWithColumns(len(buttons), buttons...)),
	)
}

func actionGroupVertical(buttons ...fyne.CanvasObject) fyne.CanvasObject {
	items := make([]fyne.CanvasObject, 0, len(buttons)*2)
	for i, button := range buttons {
		if i > 0 {
			items = append(items, verticalGap(4))
		}
		items = append(items, button)
	}
	card := canvas.NewRectangle(toolbarCardColor)
	card.CornerRadius = 20
	return container.NewStack(
		card,
		container.NewPadded(container.NewVBox(items...)),
	)
}

func LocalModeLabel(mode model.SessionMode) string {
	switch mode {
	case model.SessionModeShortBreak:
		return "短休息"
	case model.SessionModeLongBreak:
		return "长休息"
	default:
		return "工作中"
	}
}

func localModeLabel(mode model.SessionMode) string {
	return LocalModeLabel(mode)
}

func localStatusLabel(status timer.Status) string {
	switch status {
	case timer.StatusRunning:
		return "进行中"
	case timer.StatusPaused:
		return "已暂停"
	default:
		return "待开始"
	}
}

func phaseHint(snapshot timer.Snapshot) string {
	switch snapshot.Mode {
	case model.SessionModeShortBreak:
		return "短暂放松一下，再继续保持节奏"
	case model.SessionModeLongBreak:
		return "长休时间，彻底离开桌面片刻"
	default:
		if snapshot.Status == timer.StatusRunning {
			return "保持专注，这一段时间只做一件事"
		}
		return "准备进入一段专注时刻"
	}
}

func backgroundForMode(mode model.SessionMode) color.Color {
	switch mode {
	case model.SessionModeShortBreak:
		return nordPanel
	case model.SessionModeLongBreak:
		return nordPanelAlt
	default:
		return appBackgroundColor
	}
}

func formatClock(seconds int) string {
	if seconds < 0 {
		seconds = 0
	}
	return fmt.Sprintf("%02d:%02d", seconds/60, seconds%60)
}

func formatDuration(seconds int) string {
	if seconds < 0 {
		seconds = 0
	}
	return fmt.Sprintf("%02d:%02d", seconds/3600, (seconds%3600)/60)
}

func colorRGB(r, g, b uint8) color.Color {
	return &color.NRGBA{R: r, G: g, B: b, A: 255}
}

func colorTransparent() color.Color {
	return &color.NRGBA{A: 0}
}

func verticalGap(height float32) fyne.CanvasObject {
	rect := canvas.NewRectangle(colorTransparent())
	rect.SetMinSize(fyne.NewSize(1, height))
	return rect
}
