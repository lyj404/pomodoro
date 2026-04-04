package ui

import (
	"fmt"
	"image/color"
	"sync"

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
	themeBtn           *widget.Button
	langBtn            *widget.Button
	rootContent        *fyne.Container
	scroll             *container.Scroll
	layoutTier         layoutTier
	firstRender        bool
	prevSnapshot       timer.Snapshot
	prevTotalSeconds   int
	callbacks          MainCallbacks
	cachedSpacer       fyne.CanvasObject
	cachedGaps         [5]fyne.CanvasObject
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
	OnToggleTheme  func()
	OnToggleLang   func(string)
}

func (v *MainView) RefreshColors() {
	v.background.FillColor = appBackgroundColor
	v.background.Refresh()

	v.timerCard.FillColor = cardBackgroundColor
	v.timerCard.Refresh()

	v.timeLabel.Color = secondaryTextColor
	v.timeLabel.Refresh()
	v.statusLabel.Color = secondaryTextColor
	v.statusLabel.Refresh()
	v.phaseHintLabel.Color = mutedTextColor
	v.phaseHintLabel.Refresh()
	v.focusCountValue.Color = secondaryTextColor
	v.focusCountValue.Refresh()
	v.focusDurationValue.Color = secondaryTextColor
	v.focusDurationValue.Refresh()
	v.streakValue.Color = secondaryTextColor
	v.streakValue.Refresh()

	v.startBtn.RefreshColors(primaryButtonColor, primaryButtonTextColor)
	v.pauseBtn.RefreshColors(pauseButtonColor, pauseButtonTextColor)
	v.resetBtn.RefreshColors(secondaryButtonColor, secondaryTextColor)
	v.skipBtn.RefreshColors(secondaryButtonColor, secondaryTextColor)

	v.themeBtn.SetIcon(ThemeToggleIcon())
	v.themeBtn.Refresh()
	v.langBtn.SetIcon(LanguageToggleIcon())
	v.langBtn.Refresh()

	v.settingsBtn.Refresh()
	v.historyBtn.Refresh()

	v.layoutTier = layoutTierUnknown
	v.ensureLayout(v.window.Canvas().Size().Width)
}

func (v *MainView) RefreshText() {
	v.modeLabel.Text = Tr("mode.work")
	v.modeLabel.Refresh()
	v.statusLabel.Text = Tr("status.ready")
	v.statusLabel.Refresh()
	v.phaseHintLabel.Text = Tr("hint.ready")
	v.phaseHintLabel.Refresh()

	v.startBtn.RefreshText(Tr("btn.start_focus"))
	v.pauseBtn.RefreshText(Tr("btn.pause"))
	v.resetBtn.RefreshText(Tr("btn.reset"))
	v.skipBtn.RefreshText(Tr("btn.skip"))

	v.layoutTier = layoutTierUnknown
	v.ensureLayout(v.window.Canvas().Size().Width)
}

func NewMainView(win fyne.Window, callbacks MainCallbacks) *MainView {
	view := &MainView{
		window:             win,
		background:         canvas.NewRectangle(appBackgroundColor),
		accentBar:          canvas.NewRectangle(workAccentColor),
		timerCard:          canvas.NewRectangle(cardBackgroundColor),
		modeLabel:          canvas.NewText(Tr("mode.work"), workAccentColor),
		timeLabel:          canvas.NewText("25:00", secondaryTextColor),
		statusLabel:        canvas.NewText(Tr("status.ready"), secondaryTextColor),
		phaseHintLabel:     canvas.NewText(Tr("hint.ready"), mutedTextColor),
		focusCountValue:    canvas.NewText("0", secondaryTextColor),
		focusDurationValue: canvas.NewText("00:00", secondaryTextColor),
		streakValue:        canvas.NewText("0", secondaryTextColor),
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

	view.startBtn = NewActionTile(Tr("btn.start_focus"), theme.MediaPlayIcon(), primaryButtonColor, primaryButtonTextColor, callbacks.OnStart)
	view.pauseBtn = NewActionTile(Tr("btn.pause"), theme.MediaPauseIcon(), pauseButtonColor, pauseButtonTextColor, callbacks.OnPause)
	view.resetBtn = NewActionTile(Tr("btn.reset"), theme.ViewRefreshIcon(), secondaryButtonColor, secondaryTextColor, callbacks.OnReset)
	view.skipBtn = NewActionTile(Tr("btn.skip"), theme.MediaSkipNextIcon(), secondaryButtonColor, secondaryTextColor, callbacks.OnSkip)
	view.settingsBtn = widget.NewButtonWithIcon("", theme.SettingsIcon(), callbacks.OnOpenSettings)
	view.settingsBtn.Importance = widget.LowImportance
	view.historyBtn = widget.NewButtonWithIcon("", theme.HistoryIcon(), callbacks.OnOpenHistory)
	view.historyBtn.Importance = widget.LowImportance
	view.themeBtn = widget.NewButtonWithIcon("", ThemeToggleIcon(), callbacks.OnToggleTheme)
	view.themeBtn.Importance = widget.LowImportance
	view.langBtn = widget.NewButtonWithIcon("", LanguageToggleIcon(), func() {
		view.ShowLanguageDialog(func(lang string) {
			callbacks.OnToggleLang(lang)
		})
	})
	view.langBtn.Importance = widget.LowImportance

	view.rootContent = container.NewVBox()
	view.scroll = container.NewVScroll(container.NewPadded(view.rootContent))

	view.cachedSpacer = layout.NewSpacer()
	view.cachedGaps = [5]fyne.CanvasObject{
		verticalGap(1),
		verticalGap(2),
		verticalGap(3),
		verticalGap(4),
		verticalGap(8),
	}

	toolbar := container.NewHBox(
		view.historyBtn,
		layout.NewSpacer(),
		view.langBtn,
		view.themeBtn,
		view.settingsBtn,
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
		v.cachedSpacer,
		v.modeLabel,
		v.cachedGaps[0],
		v.timeLabel,
		v.cachedGaps[0],
		v.statusLabel,
		v.cachedGaps[0],
		v.phaseHintLabel,
		v.cachedSpacer,
	)

	timerCard := container.NewStack(v.timerCard, container.NewPadded(timerInner))

	statsCards := []fyne.CanvasObject{
		statCard(Tr("stat.pomodoros"), v.focusCountValue, Tr("stat.done")),
		statCard(Tr("stat.focus_time"), v.focusDurationValue, Tr("stat.work_total")),
		statCard(Tr("stat.streak"), v.streakValue, Tr("stat.in_a_row")),
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
			container.NewVBox(statsCards[0], v.cachedGaps[3], statsCards[1]),
			statsCards[2],
		)
	case layoutTierCompact:
		primaryActions = container.NewGridWithColumns(2, v.startBtn, v.pauseBtn)
		secondaryActions = actionGroup(v.resetBtn, v.skipBtn)
		stats = container.NewGridWithColumns(2,
			statsCards[0],
			statsCards[1],
			statsCards[2],
			v.cachedGaps[0],
		)
	default:
		primaryActions = container.NewVBox(v.startBtn, v.cachedGaps[3], v.pauseBtn)
		secondaryActions = actionGroupVertical(v.resetBtn, v.skipBtn)
		stats = container.NewVBox(
			statsCards[0],
			v.cachedGaps[3],
			statsCards[1],
			v.cachedGaps[3],
			statsCards[2],
		)
	}

	v.rootContent.Objects = []fyne.CanvasObject{
		v.cachedGaps[1],
		timerCard,
		v.cachedGaps[3],
		primaryActions,
		v.cachedGaps[1],
		secondaryActions,
		v.cachedGaps[2],
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
	header := dialogHeader(Tr("dialog.phase_complete"), fmt.Sprintf("%s 已结束", LocalModeLabel(snapshot.Mode)), accentColorForMode(snapshot.Mode))
	bodyText := canvas.NewText(Tr("dialog.phase_hint"), mutedTextColor)
	bodyText.TextSize = 13

	content := container.NewPadded(container.NewVBox(
		header,
		verticalGap(8),
		bodyText,
	))

	finishedDialog := dialog.NewCustom(Tr("dialog.phase_complete"), Tr("dialog.ok"), content, v.window)
	finishedDialog.Resize(fyne.NewSize(440, 220))
	finishedDialog.Show()
}

func (v *MainView) ShowError(err error) {
	dialog.ShowError(err, v.window)
}

func (v *MainView) ShowLanguageDialog(onSelect func(string)) {
	currentLang := GetLang()

	headerTitle := canvas.NewText(Tr("select_language"), secondaryTextColor)
	headerTitle.TextSize = 18
	headerTitle.TextStyle = fyne.TextStyle{Bold: true}
	headerTitle.Alignment = fyne.TextAlignCenter

	headerIcon := canvas.NewImageFromResource(LanguageToggleIcon())
	headerIcon.SetMinSize(fyne.NewSize(32, 32))

	header := container.NewVBox(
		container.NewCenter(headerIcon),
		verticalGap(8),
		headerTitle,
	)

	radioGroup := widget.NewRadioGroup([]string{Tr("simplified_chinese"), "English"}, func(s string) {})
	if currentLang == "zh" {
		radioGroup.Selected = Tr("simplified_chinese")
	} else {
		radioGroup.Selected = "English"
	}

	var langPopup *widget.PopUp

	confirmBtn := widget.NewButton(Tr("confirm"), func() {
		lang := "zh"
		if radioGroup.Selected == Tr("english") {
			lang = "en"
		}
		onSelect(lang)
		if langPopup != nil {
			langPopup.Hide()
		}
	})
	confirmBtn.Importance = widget.MediumImportance

	cancelBtn := widget.NewButton(Tr("cancel"), func() {
		if langPopup != nil {
			langPopup.Hide()
		}
	})
	cancelBtn.Importance = widget.LowImportance

	btnRow := container.NewHBox(
		layout.NewSpacer(),
		cancelBtn,
		layout.NewSpacer(),
		confirmBtn,
		layout.NewSpacer(),
	)

	mainCard := canvas.NewRectangle(cardBackgroundColor)
	mainCard.CornerRadius = 20

	content := container.NewStack(
		mainCard,
		container.NewPadded(container.NewVBox(
			verticalGap(16),
			header,
			verticalGap(12),
			container.NewCenter(radioGroup),
			verticalGap(16),
			btnRow,
			verticalGap(8),
		)),
	)

	canvasSize := v.window.Canvas().Size()
	popupSize := fyne.NewSize(300, 280)
	langPopup = widget.NewPopUp(content, v.window.Canvas())
	langPopup.Resize(popupSize)
	langPopup.Move(fyne.NewPos((canvasSize.Width-popupSize.Width)/2, (canvasSize.Height-popupSize.Height)/2))
	langPopup.Show()
}

func statCard(title string, value *canvas.Text, hint string) fyne.CanvasObject {
	card := canvas.NewRectangle(cardBackgroundColor)
	card.CornerRadius = 20

	titleText := canvas.NewText(title, mutedTextColor)
	titleText.TextSize = 11
	titleText.Alignment = fyne.TextAlignCenter

	hintText := canvas.NewText(hint, mutedTextColor)
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
		return Tr("mode.short_break")
	case model.SessionModeLongBreak:
		return Tr("mode.long_break")
	default:
		return Tr("mode.work")
	}
}

func localModeLabel(mode model.SessionMode) string {
	return LocalModeLabel(mode)
}

func localStatusLabel(status timer.Status) string {
	switch status {
	case timer.StatusRunning:
		return Tr("status.running")
	case timer.StatusPaused:
		return Tr("status.paused")
	default:
		return Tr("status.ready")
	}
}

func phaseHint(snapshot timer.Snapshot) string {
	switch snapshot.Mode {
	case model.SessionModeShortBreak:
		return Tr("hint.short_break")
	case model.SessionModeLongBreak:
		return Tr("hint.long_break")
	default:
		if snapshot.Status == timer.StatusRunning {
			return Tr("hint.stay_focused")
		}
		return Tr("准备进入一段专注时刻")
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

var clockBufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 5)
		return &buf
	},
}

func formatClock(seconds int) string {
	if seconds < 0 {
		seconds = 0
	}
	min := seconds / 60
	sec := seconds % 60
	bufp := clockBufferPool.Get().(*[]byte)
	buf := *bufp
	buf[0] = byte('0' + min/10)
	buf[1] = byte('0' + min%10)
	buf[2] = ':'
	buf[3] = byte('0' + sec/10)
	buf[4] = byte('0' + sec%10)
	s := string(buf)
	clockBufferPool.Put(bufp)
	return s
}

var durationBufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 5)
		return &buf
	},
}

func formatDuration(seconds int) string {
	if seconds < 0 {
		seconds = 0
	}
	hour := seconds / 3600
	min := (seconds % 3600) / 60
	bufp := durationBufferPool.Get().(*[]byte)
	buf := *bufp
	buf[0] = byte('0' + hour/10)
	buf[1] = byte('0' + hour%10)
	buf[2] = ':'
	buf[3] = byte('0' + min/10)
	buf[4] = byte('0' + min%10)
	s := string(buf)
	durationBufferPool.Put(bufp)
	return s
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
