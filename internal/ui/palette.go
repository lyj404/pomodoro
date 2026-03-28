package ui

import (
	"image/color"

	"github.com/lyj404/pomodoro/internal/model"
)

var (
	nordBackground = colorRGB(46, 52, 64) // #2E3440
	nordPanel      = colorRGB(59, 66, 82) // #3B4252
	nordPanelAlt   = colorRGB(67, 76, 94)
	nordPanelMuted = colorRGB(76, 86, 106)
	nordText       = colorRGB(236, 239, 244) // #ECEFF4
	nordSubtext    = colorRGB(136, 150, 168) // darker #8898AC
	nordFocus      = colorRGB(191, 97, 106)  // #BF616A
	nordBreak      = colorRGB(163, 190, 140) // #A3BE8C
	nordHighlight  = colorRGB(136, 192, 208) // #88C0D0
	nordDanger     = colorRGB(191, 97, 106)
)

var (
	appBackgroundColor      = nordBackground
	cardBackgroundColor     = nordPanel
	mutedTextColor          = nordSubtext
	workAccentColor         = nordFocus
	shortBreakAccentColor   = nordBreak
	longBreakAccentColor    = nordHighlight
	secondaryButtonColor    = nordPanelAlt
	toolbarCardColor        = nordPanel
	secondaryTextColor      = nordText
	primaryButtonColor      = nordHighlight
	primaryButtonTextColor  = nordText
	pauseButtonColor        = nordPanelMuted
	pauseButtonTextColor    = nordText
	disabledButtonColor     = nordPanelMuted
	disabledButtonTextColor = nordSubtext
)

func accentColorForMode(mode model.SessionMode) color.Color {
	switch mode {
	case model.SessionModeShortBreak:
		return shortBreakAccentColor
	case model.SessionModeLongBreak:
		return longBreakAccentColor
	default:
		return workAccentColor
	}
}
