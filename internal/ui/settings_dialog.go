package ui

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

func ShowSettingsDialog(win fyne.Window, settings model.Settings, onSave func(model.Settings)) {
	dialogWidth := float32(400)
	dialogHeight := float32(460)

	workEntry := styledEntry(strconv.Itoa(settings.WorkMinutes), workMinutesMin, workMinutesMax)
	shortBreakEntry := styledEntry(strconv.Itoa(settings.ShortBreakMinutes), shortBreakMin, shortBreakMax)
	longBreakEntry := styledEntry(strconv.Itoa(settings.LongBreakMinutes), longBreakMin, longBreakMax)
	intervalEntry := styledEntry(strconv.Itoa(settings.LongBreakInterval), breakIntervalMin, breakIntervalMax)

	autoStart := widget.NewCheck("自动开始下一阶段", nil)
	autoStart.SetChecked(settings.AutoStartNextPhase)

	soundEnabled := widget.NewCheck("开启声音提醒", nil)
	soundEnabled.SetChecked(settings.SoundEnabled)

	formCard := formSection(
		numberField("工作时长（分钟）", "范围 1-180，建议 20-60", workEntry, workMinutesMin, workMinutesMax),
		numberField("短休时长（分钟）", "范围 1-60，建议 3-10", shortBreakEntry, shortBreakMin, shortBreakMax),
		numberField("长休时长（分钟）", "范围 1-120，建议 10-30", longBreakEntry, longBreakMin, longBreakMax),
		numberField("长休触发周期", "范围 1-12，建议 4", intervalEntry, breakIntervalMin, breakIntervalMax),
		container.NewPadded(autoStart),
		container.NewPadded(soundEnabled),
	)

	content := centeredDialogContent(dialogWidth-40, formCard)

	confirm := dialog.NewCustomConfirm("设置", "保存", "取消", content, func(ok bool) {
		if !ok {
			return
		}
		if err := validateEntries(
			map[string]*widget.Entry{
				"工作时长": workEntry,
				"短休时长": shortBreakEntry,
				"长休时长": longBreakEntry,
				"长休周期": intervalEntry,
			},
		); err != nil {
			dialog.ShowError(err, win)
			return
		}

		next, err := parseSettings(
			workEntry.Text,
			shortBreakEntry.Text,
			longBreakEntry.Text,
			intervalEntry.Text,
			autoStart.Checked,
			soundEnabled.Checked,
		)
		if err != nil {
			dialog.ShowError(err, win)
			return
		}

		onSave(next)
	}, win)
	confirm.Resize(fyne.NewSize(dialogWidth, dialogHeight))
	confirm.Show()
}

func formSection(children ...fyne.CanvasObject) fyne.CanvasObject {
	card := canvas.NewRectangle(cardBackgroundColor)
	card.CornerRadius = 20
	return container.NewStack(card, container.NewPadded(container.NewVBox(children...)))
}

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

func parseSettings(work, shortBreak, longBreak, interval string, autoStart, soundEnabled bool) (model.Settings, error) {
	workMinutes, err := intInRange(work, workMinutesMin, workMinutesMax)
	if err != nil {
		return model.Settings{}, fmt.Errorf("工作时长无效: %w", err)
	}
	shortBreakMinutes, err := intInRange(shortBreak, shortBreakMin, shortBreakMax)
	if err != nil {
		return model.Settings{}, fmt.Errorf("短休时长无效: %w", err)
	}
	longBreakMinutes, err := intInRange(longBreak, longBreakMin, longBreakMax)
	if err != nil {
		return model.Settings{}, fmt.Errorf("长休时长无效: %w", err)
	}
	longBreakInterval, err := intInRange(interval, breakIntervalMin, breakIntervalMax)
	if err != nil {
		return model.Settings{}, fmt.Errorf("长休周期无效: %w", err)
	}

	return model.Settings{
		WorkMinutes:        workMinutes,
		ShortBreakMinutes:  shortBreakMinutes,
		LongBreakMinutes:   longBreakMinutes,
		LongBreakInterval:  longBreakInterval,
		AutoStartNextPhase: autoStart,
		SoundEnabled:       soundEnabled,
	}, nil
}

func intInRange(value string, min, max int) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("请输入整数（%d-%d）", min, max)
	}
	if parsed < min || parsed > max {
		return 0, fmt.Errorf("请输入 %d-%d 之间的整数", min, max)
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
