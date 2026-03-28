package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/lyj404/pomodoro/internal/model"
)

func ShowHistoryDialog(
	win fyne.Window,
	sessions []model.Session,
	onDeleteMany func([]int64) error,
	onRefresh func() ([]model.Session, error),
	onDeleted func(),
) {
	selected := map[int64]bool{}

	selectedCount := canvas.NewText("已选 0 条", mutedTextColor)
	selectedCount.TextSize = 13

	listHost := container.NewMax()
	var currentSessions []model.Session

	var refreshList func(current []model.Session)

	selectAllCheck := widget.NewCheck("全选", func(checked bool) {
		if checked {
			for _, session := range currentSessions {
				selected[session.ID] = true
			}
		} else {
			for id := range selected {
				delete(selected, id)
			}
		}
		refreshList(currentSessions)
	})

	refreshList = func(current []model.Session) {
		currentSessions = current
		selectedCount.Text = fmt.Sprintf("已选 %d 条", len(selectedIDs(selected)))
		selectedCount.Refresh()
		allSelected := len(current) > 0 && len(selectedIDs(selected)) == len(current)
		selectAllCheck.SetChecked(allSelected)
		listHost.Objects = []fyne.CanvasObject{buildHistoryBody(current, selected, func() {
			selectedCount.Text = fmt.Sprintf("已选 %d 条", len(selectedIDs(selected)))
			selectedCount.Refresh()
			allSelected := len(current) > 0 && len(selectedIDs(selected)) == len(current)
			selectAllCheck.SetChecked(allSelected)
		})}
		listHost.Refresh()
	}

	deleteSelectedBtn := NewActionTile("删除选中", nil, nordDanger, nordText, func() {
		ids := selectedIDs(selected)
		if len(ids) == 0 {
			dialog.ShowInformation("提示", "请先选择至少一条记录。", win)
			return
		}

		dialog.NewConfirm("批量删除", fmt.Sprintf("确认删除选中的 %d 条记录？", len(ids)), func(ok bool) {
			if !ok {
				return
			}
			if err := onDeleteMany(ids); err != nil {
				dialog.ShowError(err, win)
				return
			}

			for _, id := range ids {
				delete(selected, id)
			}

			nextSessions, err := onRefresh()
			if err != nil {
				dialog.ShowError(err, win)
				return
			}

			refreshList(nextSessions)
			onDeleted()
		}, win).Show()
	})

	toolbar := container.NewVBox(
		selectedCount,
		verticalGap(4),
		container.NewGridWithColumns(2, selectAllCheck, deleteSelectedBtn),
	)

	refreshList(sessions)
	scroll := container.NewVScroll(container.NewHBox(
		listHost,
	))
	scroll.SetMinSize(fyne.NewSize(340, 180))
	listCard := formSection(
		toolbar,
		verticalGap(4),
		scroll,
	)

	content := container.NewPadded(centeredDialogContent(380, listCard))

	historyDialog := dialog.NewCustom("历史记录", "关闭", content, win)
	historyDialog.Resize(fyne.NewSize(440, 340))
	historyDialog.Show()
}

func buildHistoryBody(
	sessions []model.Session,
	selected map[int64]bool,
	onSelectionChanged func(),
) fyne.CanvasObject {
	if len(sessions) == 0 {
		return emptyHistoryCard()
	}

	items := make([]fyne.CanvasObject, 0, len(sessions))
	for _, session := range sessions {
		items = append(items, historyRow(session, selected, onSelectionChanged))
	}
	return container.NewVBox(items...)
}

func historyRow(
	session model.Session,
	selected map[int64]bool,
	onSelectionChanged func(),
) fyne.CanvasObject {
	card := canvas.NewRectangle(historyCardColor(selected[session.ID]))
	card.CornerRadius = 12

	border := canvas.NewRectangle(historyBorderColor(selected[session.ID], session.Mode))
	border.CornerRadius = 12

	modeText := canvas.NewText(localModeLabel(session.Mode), accentColorForMode(session.Mode))
	modeText.TextSize = 14
	modeText.TextStyle = fyne.TextStyle{Bold: true}

	statusText := canvas.NewText(sessionStatusText(session.Completed), statusColor(session.Completed))
	statusText.TextSize = 11
	statusText.TextStyle = fyne.TextStyle{Bold: true}

	metaText := canvas.NewText(
		fmt.Sprintf(
			"开始: %s",
			session.StartedAt.Format("2006-01-02 15:04"),
		),
		nordSubtext,
	)
	metaText.TextSize = 11

	plannedText := canvas.NewText(fmt.Sprintf("计划: %s", formatClock(session.PlannedSeconds)), nordSubtext)
	plannedText.TextSize = 11
	actualText := canvas.NewText(fmt.Sprintf("实际: %s", formatClock(session.ActualSeconds)), nordSubtext)
	actualText.TextSize = 11

	check := widget.NewCheck("", func(checked bool) {
		if checked {
			selected[session.ID] = true
		} else {
			delete(selected, session.ID)
		}
		card.FillColor = historyCardColor(checked)
		border.FillColor = historyBorderColor(checked, session.Mode)
		card.Refresh()
		border.Refresh()
		onSelectionChanged()
	})
	check.SetChecked(selected[session.ID])

	details := container.NewGridWithColumns(2,
		container.NewPadded(plannedText),
		container.NewPadded(actualText),
	)
	info := container.NewVBox(
		modeText,
		metaText,
		details,
	)
	header := container.NewBorder(nil, nil, container.NewCenter(check), container.NewPadded(statusText), info)
	borderWrap := container.NewStack(card, container.NewPadded(header))
	return container.NewPadded(container.NewStack(border, borderWrap))
}

func emptyHistoryCard() fyne.CanvasObject {
	card := canvas.NewRectangle(cardBackgroundColor)
	card.CornerRadius = 20

	title := canvas.NewText("暂无记录", nordText)
	title.TextSize = 20
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := canvas.NewText("开始一次专注后，这里会自动出现历史记录。", nordSubtext)
	subtitle.TextSize = 13
	subtitle.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		layout.NewSpacer(),
		title,
		verticalGap(4),
		subtitle,
		layout.NewSpacer(),
	)

	return container.NewStack(card, container.NewPadded(content))
}

func selectedIDs(selected map[int64]bool) []int64 {
	ids := make([]int64, 0, len(selected))
	for id, checked := range selected {
		if checked {
			ids = append(ids, id)
		}
	}
	return ids
}

func sessionStatusText(completed bool) string {
	if completed {
		return "已完成"
	}
	return "未完成"
}

func statusColor(completed bool) color.Color {
	if completed {
		return nordHighlight
	}
	return nordDanger
}

func historyCardColor(selected bool) color.Color {
	if selected {
		return nordPanelAlt
	}
	return cardBackgroundColor
}

func historyBorderColor(selected bool, mode model.SessionMode) color.Color {
	if selected {
		return accentColorForMode(mode)
	}
	return nordPanelMuted
}
