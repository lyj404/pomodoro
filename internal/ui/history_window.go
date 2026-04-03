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

	dialogWidth := float32(400)

	contentWidth := dialogWidth - 40
	scrollWidth := contentWidth
	compactRows := dialogWidth < 480

	selectedCount := canvas.NewText(Tr("selected")+": 0 "+Tr("items"), secondaryTextColor)
	selectedCount.TextSize = 14

	listHost := container.NewMax()
	var currentSessions []model.Session
	var syncingSelectAll bool
	var refreshList func(current []model.Session)
	var loadPage func(offset int)

	pageInfo := canvas.NewText("", secondaryTextColor)
	pageInfo.TextSize = 13

	prevBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), nil)
	prevBtn.Importance = widget.LowImportance
	nextBtn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), nil)
	nextBtn.Importance = widget.LowImportance

	selectAllCheck := widget.NewCheck(Tr("select_all"), func(checked bool) {
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

	refreshList = func(current []model.Session) {
		currentSessions = current
		selectedInView := selectedIDsInSessions(selected, current)
		selectedCount.Text = fmt.Sprintf("已选: %d 条", len(selectedInView))
		selectedCount.Refresh()
		allSelected := len(current) > 0 && len(selectedInView) == len(current)
		syncingSelectAll = true
		selectAllCheck.SetChecked(allSelected)
		if len(current) == 0 {
			selectAllCheck.Disable()
		} else {
			selectAllCheck.Enable()
		}
		syncingSelectAll = false

		currentPage := currentOffset/historyPageSize + 1
		totalPages := (total + historyPageSize - 1) / historyPageSize
		if totalPages == 0 {
			totalPages = 1
		}
		pageInfo.Text = fmt.Sprintf("%d/%d (共 %d 条)", currentPage, totalPages, total)
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

		listHost.Objects = []fyne.CanvasObject{buildHistoryBody(current, selected, compactRows, scrollWidth, func() {
			selectedInView := selectedIDsInSessions(selected, current)
			selectedCount.Text = fmt.Sprintf("%s: %d %s", Tr("selected"), len(selectedInView), Tr("items"))
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

	deleteSelectedBtn := NewActionTile(Tr("delete_selected"), nil, nordDanger, secondaryTextColor, func() {
		ids := selectedIDsInSessions(selected, currentSessions)
		if len(ids) == 0 {
			dialog.ShowInformation(Tr("notice"), Tr("select_records_hint"), win)
			return
		}

		dialog.NewConfirm(Tr("batch_delete"), fmt.Sprintf(Tr("confirm_delete_hint"), len(ids)), func(ok bool) {
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

	paginationRow := container.NewHBox(
		prevBtn,
		pageInfo,
		nextBtn,
	)

	headerRow := container.NewBorder(nil, nil, nil, paginationRow, selectedCount)
	// 操作行：全选和删除分布两头
	actionRow := container.NewHBox(
		selectAllCheck,
		layout.NewSpacer(), // 把删除按钮推到最右边
		deleteSelectedBtn,
	)
	toolbar := container.NewVBox(
		headerRow,
		verticalGap(8),
		actionRow,
	)
	// 列表区域
	scroll := container.NewVScroll(listHost)
	scroll.SetMinSize(fyne.NewSize(340, 280))
	// 核心卡片：包含工具栏和滚动列表
	listCard := formSection(
		toolbar,
		verticalGap(10),
		scroll,
	)
	// 最终包装：不再使用强制 360 宽度的 centeredDialogContent
	content := container.NewPadded(listCard)
	historyDialog := dialog.NewCustom(Tr("history"), Tr("close"), content, win)
	historyDialog.Resize(fyne.NewSize(400, 520)) // 稍微增加高度以容纳空状态
	refreshList(sessions)
	historyDialog.Show()
}

func buildHistoryBody(
	sessions []model.Session,
	selected map[int64]bool,
	compact bool,
	width float32,
	onSelectionChanged func(),
) fyne.CanvasObject {
	if len(sessions) == 0 {
		return emptyHistoryCard()
	}

	items := make([]fyne.CanvasObject, 0, len(sessions))
	for i, session := range sessions {
		if i > 0 {
			items = append(items, verticalGap(6))
		}
		items = append(items, historyRow(session, selected, compact, width, onSelectionChanged))
	}
	return container.NewVBox(items...)
}

func historyRow(
	session model.Session,
	selected map[int64]bool,
	compact bool,
	rowWidth float32,
	onSelectionChanged func(),
) fyne.CanvasObject {
	modeText := canvas.NewText(localModeLabel(session.Mode), accentColorForMode(session.Mode))
	modeText.TextSize = 15
	modeText.TextStyle = fyne.TextStyle{Bold: true}

	statusText := canvas.NewText(sessionStatusText(session.Completed), statusColor(session.Completed))
	statusText.TextSize = 12
	statusText.TextStyle = fyne.TextStyle{Bold: true}
	statusText.Alignment = fyne.TextAlignTrailing

	metaText := canvas.NewText(
		fmt.Sprintf("开始: %s", session.StartedAt.Format("2006-01-02 15:04")),
		secondaryTextColor,
	)
	metaText.TextSize = 13

	plannedText := canvas.NewText(fmt.Sprintf("%s: %s", Tr("planned"), formatClock(session.PlannedSeconds)), secondaryTextColor)
	plannedText.TextSize = 13
	actualText := canvas.NewText(fmt.Sprintf("%s: %s", Tr("actual"), formatClock(session.ActualSeconds)), secondaryTextColor)
	actualText.TextSize = 13

	check := widget.NewCheck("", func(checked bool) {
		if checked {
			selected[session.ID] = true
		} else {
			delete(selected, session.ID)
		}
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

	info := container.NewVBox(modeText, metaText, details)

	rowBg := canvas.NewRectangle(historyCardColor(selected[session.ID]))
	rowBg.CornerRadius = 12

	borderBg := canvas.NewRectangle(historyBorderColor(selected[session.ID], session.Mode))
	borderBg.CornerRadius = 12

	content := container.NewBorder(nil, nil, check, statusText, info)

	rowBg.SetMinSize(fyne.NewSize(rowWidth, 80))
	borderBg.SetMinSize(fyne.NewSize(rowWidth, 80))

	inner := container.NewStack(rowBg, content)
	borderWrap := container.NewStack(borderBg, inner)

	borderWrap.Resize(fyne.NewSize(rowWidth, 80))

	return borderWrap
}

func emptyHistoryCard() fyne.CanvasObject {
	card := canvas.NewRectangle(cardBackgroundColor)
	card.CornerRadius = 20

	title := canvas.NewText(Tr("no_records"), secondaryTextColor)
	title.TextSize = 22
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := canvas.NewText(Tr("history.empty_hint"), secondaryTextColor)
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
	filter.Resize(fyne.NewSize(100, 30))
	return container.NewHBox(
		canvas.NewText(label, secondaryTextColor),
		layout.NewSpacer(),
		filter,
	)
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
	return mutedTextColor
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
