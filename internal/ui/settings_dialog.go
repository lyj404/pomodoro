package ui

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/lyj404/pomodoro/internal/model"
)

func ShowSettingsDialog(win fyne.Window, settings model.Settings, onSave func(model.Settings)) {
	workEntry := styledEntry(strconv.Itoa(settings.WorkMinutes))
	shortBreakEntry := styledEntry(strconv.Itoa(settings.ShortBreakMinutes))
	longBreakEntry := styledEntry(strconv.Itoa(settings.LongBreakMinutes))
	intervalEntry := styledEntry(strconv.Itoa(settings.LongBreakInterval))

	autoStart := widget.NewCheck("自动开始下一阶段", nil)
	autoStart.SetChecked(settings.AutoStartNextPhase)

	formCard := formSection(
		formRow("工作时长（分钟）", workEntry),
		formRow("短休时长（分钟）", shortBreakEntry),
		formRow("长休时长（分钟）", longBreakEntry),
		formRow("长休触发周期", intervalEntry),
		container.NewPadded(autoStart),
	)

	content := container.NewPadded(centeredDialogContent(380, formCard))

	confirm := dialog.NewCustomConfirm("设置", "保存", "取消", content, func(ok bool) {
		if !ok {
			return
		}

		next, err := parseSettings(
			workEntry.Text,
			shortBreakEntry.Text,
			longBreakEntry.Text,
			intervalEntry.Text,
			autoStart.Checked,
		)
		if err != nil {
			dialog.ShowError(err, win)
			return
		}

		onSave(next)
	}, win)
	confirm.Resize(fyne.NewSize(460, 390))
	confirm.Show()
}

func formSection(children ...fyne.CanvasObject) fyne.CanvasObject {
	card := canvas.NewRectangle(cardBackgroundColor)
	card.CornerRadius = 20
	return container.NewStack(card, container.NewPadded(container.NewVBox(children...)))
}

func formRow(label string, field fyne.CanvasObject) fyne.CanvasObject {
	labelText := canvas.NewText(label, nordSubtext)
	labelText.TextSize = 13
	return container.NewVBox(labelText, field)
}

func styledEntry(value string) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetText(value)
	return entry
}

func parseSettings(work, shortBreak, longBreak, interval string, autoStart bool) (model.Settings, error) {
	workMinutes, err := positiveInt(work)
	if err != nil {
		return model.Settings{}, fmt.Errorf("工作时长无效: %w", err)
	}
	shortBreakMinutes, err := positiveInt(shortBreak)
	if err != nil {
		return model.Settings{}, fmt.Errorf("短休时长无效: %w", err)
	}
	longBreakMinutes, err := positiveInt(longBreak)
	if err != nil {
		return model.Settings{}, fmt.Errorf("长休时长无效: %w", err)
	}
	longBreakInterval, err := positiveInt(interval)
	if err != nil {
		return model.Settings{}, fmt.Errorf("长休周期无效: %w", err)
	}

	return model.Settings{
		WorkMinutes:        workMinutes,
		ShortBreakMinutes:  shortBreakMinutes,
		LongBreakMinutes:   longBreakMinutes,
		LongBreakInterval:  longBreakInterval,
		AutoStartNextPhase: autoStart,
	}, nil
}

func positiveInt(value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("请输入大于 0 的整数")
	}
	return parsed, nil
}
