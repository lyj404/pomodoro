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
)

const (
	historyModeAll   = "全部模式"
	historyModeWork  = "工作"
	historyModeShort = "短休息"
	historyModeLong  = "长休息"

	historyStatusAll        = "全部状态"
	historyStatusCompleted  = "已完成"
	historyStatusIncomplete = "未完成"

	historyPageSize = 20
)

func ShowHistoryDialog(
	win fyne.Window,
	sessions []model.Session,
	totalCount int,
	onDeleteMany func([]int64) error,
	onRefresh func(offset int) ([]model.Session, int, error),
	onDeleted func(),
) {
	selected := map[int64]bool{}
	currentOffset := 0
	total := totalCount
	windowSize := win.Canvas().Size()
	dialogWidth := clampFloat32(windowSize.Width-40, 380, 560)
	dialogHeight := clampFloat32(windowSize.Height-40, 380, 520)
	contentWidth := clampFloat32(dialogWidth-40, 340, 520)
	scrollWidth := clampFloat32(dialogWidth-100, 280, 460)
	scrollHeight := clampFloat32(dialogHeight-280, 160, 280)
	compactRows := dialogWidth < 500

	selectedCount := canvas.NewText("当前筛选已选 0 条", nordText)
	selectedCount.TextSize = 14

	listHost := container.NewMax()
	var currentSessions []model.Session
	var syncingSelectAll bool
	var refreshList func(current []model.Session)
	var loadPage func(offset int)

	pageInfo := canvas.NewText("", nordText)
	pageInfo.TextSize = 13

	prevBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), nil)
	prevBtn.Importance = widget.LowImportance
	nextBtn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), nil)
	nextBtn.Importance = widget.LowImportance

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

		currentPage := currentOffset/historyPageSize + 1
		totalPages := (total + historyPageSize - 1) / historyPageSize
		if totalPages == 0 {
			totalPages = 1
		}
		pageInfo.Text = fmt.Sprintf("第 %d/%d 页 (共 %d 条)", currentPage, totalPages, total)
		pageInfo.Refresh()

		if currentOffset == 0 {
			prevBtn.Hide()
		} else {
			prevBtn.Show()
		}
		if currentOffset+len(current) >= total {
			nextBtn.Hide()
		} else {
			nextBtn.Show()
		}

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

	loadPage = func(offset int) {
		currentOffset = offset
		selected = map[int64]bool{}
		sessions, totalCount, err := onRefresh(offset)
		total = totalCount
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		refreshList(sessions)
	}

	modeFilter.OnChanged = func(string) {
		loadPage(0)
	}
	statusFilter.OnChanged = func(string) {
		loadPage(0)
	}

	prevBtn.OnTapped = func() {
		if currentOffset > 0 {
			loadPage(currentOffset - historyPageSize)
		}
	}

	nextBtn.OnTapped = func() {
		if currentOffset+len(currentSessions) < total {
			loadPage(currentOffset + historyPageSize)
		}
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

			onDeleted()
			loadPage(currentOffset)
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

	paginationControls := container.NewHBox(
		prevBtn,
		container.NewCenter(pageInfo),
		nextBtn,
	)

	toolbar := container.NewVBox(
		selectedCount,
		verticalGap(4),
		filterControls,
		verticalGap(3),
		actionControls,
		verticalGap(3),
		paginationControls,
	)

	refreshList(sessions)
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
	modeText.TextSize = 15
	modeText.TextStyle = fyne.TextStyle{Bold: true}

	statusText := canvas.NewText(sessionStatusText(session.Completed), statusColor(session.Completed))
	statusText.TextSize = 12
	statusText.TextStyle = fyne.TextStyle{Bold: true}

	metaText := canvas.NewText(
		fmt.Sprintf(
			"开始: %s",
			session.StartedAt.Format("2006-01-02 15:04"),
		),
		nordText,
	)
	metaText.TextSize = 13

	plannedText := canvas.NewText(fmt.Sprintf("计划: %s", formatClock(session.PlannedSeconds)), nordText)
	plannedText.TextSize = 13
	actualText := canvas.NewText(fmt.Sprintf("实际: %s", formatClock(session.ActualSeconds)), nordText)
	actualText.TextSize = 13

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
	title.TextSize = 22
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := canvas.NewText("开始一次专注后，这里会自动出现历史记录。", nordText)
	subtitle.TextSize = 14
	subtitle.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		layout.NewSpacer(),
		title,
		verticalGap(8),
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
	title := canvas.NewText(label, nordText)
	title.TextSize = 13
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
