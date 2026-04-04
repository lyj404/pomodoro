package ui

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/lyj404/pomodoro/internal/model"
)

const (
	workMinutesMin   = 1
	workMinutesMax   = 180
	shortBreakMin    = 1
	shortBreakMax    = 60
	longBreakMin     = 1
	longBreakMax     = 120
	breakIntervalMin = 1
	breakIntervalMax = 12
)

func ShowSettingsPopup(canv fyne.Canvas, settings model.Settings, onSave func(model.Settings)) {
	workEntry := styledEntry(strconv.Itoa(settings.WorkMinutes), workMinutesMin, workMinutesMax)
	shortBreakEntry := styledEntry(strconv.Itoa(settings.ShortBreakMinutes), shortBreakMin, shortBreakMax)
	longBreakEntry := styledEntry(strconv.Itoa(settings.LongBreakMinutes), longBreakMin, longBreakMax)
	intervalEntry := styledEntry(strconv.Itoa(settings.LongBreakInterval), breakIntervalMin, breakIntervalMax)

	autoStart := widget.NewCheck(Tr("settings.auto_start"), nil)
	autoStart.SetChecked(settings.AutoStartNextPhase)

	soundEnabled := widget.NewCheck(Tr("settings.sound"), nil)
	soundEnabled.SetChecked(settings.SoundEnabled)

	headerTitle := canvas.NewText(Tr("settings"), secondaryTextColor)
	headerTitle.TextSize = 18
	headerTitle.TextStyle = fyne.TextStyle{Bold: true}
	headerTitle.Alignment = fyne.TextAlignCenter

	closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		if settingsPopup != nil {
			settingsPopup.Hide()
		}
	})
	closeBtn.Importance = widget.LowImportance

	confirmBtn := widget.NewButton(Tr("save"), func() {
		if err := validateEntries(
			map[string]*widget.Entry{
				Tr("settings.work_minutes"):        workEntry,
				Tr("settings.short_break_minutes"): shortBreakEntry,
				Tr("settings.long_break_minutes"):  longBreakEntry,
				Tr("settings.long_break_interval"): intervalEntry,
			},
		); err != nil {
			return
		}

		next, err := parseSettings(
			workEntry.Text,
			shortBreakEntry.Text,
			longBreakEntry.Text,
			intervalEntry.Text,
			autoStart.Checked,
			soundEnabled.Checked,
			settings.Language,
		)
		if err != nil {
			return
		}

		onSave(next)
		if settingsPopup != nil {
			settingsPopup.Hide()
		}
	})
	confirmBtn.Importance = widget.MediumImportance

	cancelBtn := widget.NewButton(Tr("cancel"), func() {
		if settingsPopup != nil {
			settingsPopup.Hide()
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

	formContent := container.NewVBox(
		numberField(Tr("settings.work_minutes"), Tr("settings.work_hint"), workEntry, workMinutesMin, workMinutesMax),
		verticalGap(4),
		numberField(Tr("settings.short_break_minutes"), Tr("settings.short_hint"), shortBreakEntry, shortBreakMin, shortBreakMax),
		verticalGap(4),
		numberField(Tr("settings.long_break_minutes"), Tr("settings.long_hint"), longBreakEntry, longBreakMin, longBreakMax),
		verticalGap(4),
		numberField(Tr("settings.long_break_interval"), Tr("settings.interval_hint"), intervalEntry, breakIntervalMin, breakIntervalMax),
		verticalGap(8),
		autoStart,
		soundEnabled,
	)

	headerRow := container.NewBorder(nil, nil, nil, closeBtn, headerTitle)

	formScroll := container.NewVScroll(formContent)
	formScroll.SetMinSize(fyne.NewSize(320, 600))

	mainCard := canvas.NewRectangle(cardBackgroundColor)
	mainCard.CornerRadius = 20

	content := container.NewStack(
		mainCard,
		container.NewVBox(
			headerRow,
			formScroll,
			layout.NewSpacer(),
			btnRow,
		),
	)

	canvasSize := canv.Size()
	popupSize := fyne.NewSize(360, 640)
	settingsPopup = widget.NewPopUp(content, canv)
	settingsPopup.Resize(popupSize)
	settingsPopup.Move(fyne.NewPos((canvasSize.Width-popupSize.Width)/2, (canvasSize.Height-popupSize.Height)/2))
	settingsPopup.Show()
}

var settingsPopup *widget.PopUp

func numberField(label, hint string, entry *widget.Entry, min, max int) fyne.CanvasObject {
	labelText := canvas.NewText(label, secondaryTextColor)
	labelText.TextSize = 14

	hintText := canvas.NewText(hint, mutedTextColor)
	hintText.TextSize = 11

	stepDown := widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
		adjustEntry(entry, -1, min, max)
	})
	stepDown.Importance = widget.LowImportance

	stepUp := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		adjustEntry(entry, 1, min, max)
	})
	stepUp.Importance = widget.LowImportance

	field := container.NewBorder(nil, nil, stepDown, stepUp, entry)
	return container.NewVBox(labelText, field, hintText)
}

func styledEntry(value string, min, max int) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(fmt.Sprintf("%d-%d", min, max))
	entry.Validator = func(text string) error {
		_, err := intInRange(text, min, max)
		return err
	}
	entry.OnChanged = func(string) {
		_ = entry.Validate()
	}
	entry.SetText(value)
	return entry
}

func parseSettings(work, shortBreak, longBreak, interval string, autoStart, soundEnabled bool, language string) (model.Settings, error) {
	workMinutes, err := intInRange(work, workMinutesMin, workMinutesMax)
	if err != nil {
		return model.Settings{}, fmt.Errorf("%s: %w", Tr("settings.work_minutes"), err)
	}
	shortBreakMinutes, err := intInRange(shortBreak, shortBreakMin, shortBreakMax)
	if err != nil {
		return model.Settings{}, fmt.Errorf("%s: %w", Tr("settings.short_break_minutes"), err)
	}
	longBreakMinutes, err := intInRange(longBreak, longBreakMin, longBreakMax)
	if err != nil {
		return model.Settings{}, fmt.Errorf("%s: %w", Tr("settings.long_break_minutes"), err)
	}
	longBreakInterval, err := intInRange(interval, breakIntervalMin, breakIntervalMax)
	if err != nil {
		return model.Settings{}, fmt.Errorf("%s: %w", Tr("settings.long_break_interval"), err)
	}

	return model.Settings{
		WorkMinutes:        workMinutes,
		ShortBreakMinutes:  shortBreakMinutes,
		LongBreakMinutes:   longBreakMinutes,
		LongBreakInterval:  longBreakInterval,
		AutoStartNextPhase: autoStart,
		SoundEnabled:       soundEnabled,
		Language:           language,
	}, nil
}

func intInRange(value string, min, max int) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer (%d-%d)", min, max)
	}
	if parsed < min || parsed > max {
		return 0, fmt.Errorf("value must be between %d and %d", min, max)
	}
	return parsed, nil
}

func validateEntries(entries map[string]*widget.Entry) error {
	for label, entry := range entries {
		if err := entry.Validate(); err != nil {
			return fmt.Errorf("%s: %w", label, err)
		}
	}
	return nil
}

func adjustEntry(entry *widget.Entry, delta, min, max int) {
	current, err := strconv.Atoi(entry.Text)
	if err != nil {
		current = min
	}
	next := current + delta
	if next < min {
		next = min
	}
	if next > max {
		next = max
	}
	entry.SetText(strconv.Itoa(next))
	_ = entry.Validate()
}
