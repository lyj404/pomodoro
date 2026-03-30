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

const (
	historyModeAll   = "全部模式"
	historyModeWork  = "工作"
	historyModeShort = "短休息"
	historyModeLong  = "长休息"

	historyStatusAll        = "全部状态"
	historyStatusCompleted  = "已完成"
	historyStatusIncomplete = "未完成"
)

func ShowHistoryDialog(
	win fyne.Window,
	sessions []model.Session,
	onDeleteMany func([]int64) error,
	onRefresh func() ([]model.Session, error),
	onDeleted func(),
) {
	selected := map[int64]bool{}
	allSessions := sessions
	windowSize := win.Canvas().Size()
	dialogWidth := clampFloat32(windowSize.Width-40, 380, 560)
	dialogHeight := clampFloat32(windowSize.Height-40, 320, 460)
	contentWidth := clampFloat32(dialogWidth-40, 340, 520)
	scrollWidth := clampFloat32(dialogWidth-100, 280, 460)
	scrollHeight := clampFloat32(dialogHeight-220, 160, 260)
	compactRows := dialogWidth < 500

	selectedCount := canvas.NewText("当前筛选已选 0 条", mutedTextColor)
	selectedCount.TextSize = 13

	listHost := container.NewMax()
	var currentSessions []model.Session
	var syncingSelectAll bool
	var refreshList func(current []model.Session)
	var refreshVisibleList func()

	selectAllCheck := widget.NewCheck("全选", func(checked bool) {
		if syncingSelectAll {
			return
		}
		if checked {
			for _, session := range currentSessions {
				selected[session.ID] = true
			}
		} else {
			for _, session := range currentSessions {
				delete(selected, session.ID)
			}
		}
		refreshList(currentSessions)
	})

	modeFilter := widget.NewSelect(
		[]string{historyModeAll, historyModeWork, historyModeShort, historyModeLong},
		nil,
	)
	modeFilter.SetSelected(historyModeAll)

	statusFilter := widget.NewSelect(
		[]string{historyStatusAll, historyStatusCompleted, historyStatusIncomplete},
		nil,
	)
	statusFilter.SetSelected(historyStatusAll)

	refreshList = func(current []model.Session) {
		currentSessions = current
		selectedInView := selectedIDsInSessions(selected, current)
		selectedCount.Text = fmt.Sprintf("当前筛选已选 %d 条", len(selectedInView))
		selectedCount.Refresh()
		allSelected := len(current) > 0 && len(selectedInView) == len(current)
		syncingSelectAll = true
		selectAllCheck.SetChecked(allSelected)
		syncingSelectAll = false
		listHost.Objects = []fyne.CanvasObject{buildHistoryBody(current, selected, compactRows, func() {
			selectedInView := selectedIDsInSessions(selected, current)
			selectedCount.Text = fmt.Sprintf("当前筛选已选 %d 条", len(selectedInView))
			selectedCount.Refresh()
			allSelected := len(current) > 0 && len(selectedInView) == len(current)
			syncingSelectAll = true
			selectAllCheck.SetChecked(allSelected)
			syncingSelectAll = false
		})}
		listHost.Refresh()
	}

	refreshVisibleList = func() {
		refreshList(filterSessions(allSessions, modeFilter.Selected, statusFilter.Selected))
	}
	modeFilter.OnChanged = func(string) {
		refreshVisibleList()
	}
	statusFilter.OnChanged = func(string) {
		refreshVisibleList()
	}

	deleteSelectedBtn := NewActionTile("删除选中", nil, nordDanger, nordText, func() {
		ids := selectedIDsInSessions(selected, currentSessions)
		if len(ids) == 0 {
			dialog.ShowInformation("提示", "请先选择当前筛选结果中的记录。", win)
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
			allSessions = nextSessions
			refreshVisibleList()
			onDeleted()
		}, win).Show()
	})

	modeFilterWrap := labeledFilter("模式筛选", modeFilter)
	statusFilterWrap := labeledFilter("完成状态", statusFilter)
	var filterControls fyne.CanvasObject
	var actionControls fyne.CanvasObject
	if dialogWidth >= 460 {
		filterControls = container.NewGridWithColumns(2, modeFilterWrap, statusFilterWrap)
		actionControls = container.NewGridWithColumns(2, selectAllCheck, deleteSelectedBtn)
	} else {
		filterControls = container.NewVBox(modeFilterWrap, verticalGap(2), statusFilterWrap)
		actionControls = container.NewVBox(selectAllCheck, verticalGap(2), deleteSelectedBtn)
	}

	toolbar := container.NewVBox(
		selectedCount,
		verticalGap(4),
		filterControls,
		verticalGap(3),
		actionControls,
	)

	refreshVisibleList()
	scroll := container.NewVScroll(container.NewHBox(
		listHost,
	))
	scroll.SetMinSize(fyne.NewSize(scrollWidth, scrollHeight))
	listCard := formSection(
		toolbar,
		verticalGap(4),
		scroll,
	)

	content := container.NewPadded(centeredDialogContent(contentWidth, listCard))

	historyDialog := dialog.NewCustom("历史记录", "关闭", content, win)
	historyDialog.Resize(fyne.NewSize(dialogWidth, dialogHeight))
	historyDialog.Show()
}

func buildHistoryBody(
	sessions []model.Session,
	selected map[int64]bool,
	compact bool,
	onSelectionChanged func(),
) fyne.CanvasObject {
	if len(sessions) == 0 {
		return emptyHistoryCard()
	}

	items := make([]fyne.CanvasObject, 0, len(sessions))
	for _, session := range sessions {
		items = append(items, historyRow(session, selected, compact, onSelectionChanged))
	}
	return container.NewVBox(items...)
}

func historyRow(
	session model.Session,
	selected map[int64]bool,
	compact bool,
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

	var details fyne.CanvasObject
	if compact {
		details = container.NewVBox(plannedText, verticalGap(1), actualText)
	} else {
		details = container.NewGridWithColumns(2,
			container.NewPadded(plannedText),
			container.NewPadded(actualText),
		)
	}
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

func selectedIDsInSessions(selected map[int64]bool, sessions []model.Session) []int64 {
	ids := make([]int64, 0, len(sessions))
	for _, session := range sessions {
		if selected[session.ID] {
			ids = append(ids, session.ID)
		}
	}
	return ids
}

func labeledFilter(label string, filter *widget.Select) fyne.CanvasObject {
	title := canvas.NewText(label, nordSubtext)
	title.TextSize = 11
	return container.NewVBox(title, filter)
}

func filterSessions(sessions []model.Session, modeFilter, statusFilter string) []model.Session {
	filtered := make([]model.Session, 0, len(sessions))
	for _, session := range sessions {
		if !modeMatchesFilter(session.Mode, modeFilter) {
			continue
		}
		if !statusMatchesFilter(session.Completed, statusFilter) {
			continue
		}
		filtered = append(filtered, session)
	}
	return filtered
}

func modeMatchesFilter(mode model.SessionMode, modeFilter string) bool {
	switch modeFilter {
	case historyModeWork:
		return mode == model.SessionModeWork
	case historyModeShort:
		return mode == model.SessionModeShortBreak
	case historyModeLong:
		return mode == model.SessionModeLongBreak
	default:
		return true
	}
}

func statusMatchesFilter(completed bool, statusFilter string) bool {
	switch statusFilter {
	case historyStatusCompleted:
		return completed
	case historyStatusIncomplete:
		return !completed
	default:
		return true
	}
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

func clampFloat32(value, min, max float32) float32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
