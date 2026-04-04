package ui

import (
	"image/color"

	"github.com/lyj404/pomodoro/internal/model"
)

var (
	nordBackground   = colorRGB(46, 52, 64) // #2E3440
	nordPanel        = colorRGB(59, 66, 82) // #3B4252
	nordPanelAlt     = colorRGB(67, 76, 94)
	nordPanelMuted   = colorRGB(76, 86, 106)
	nordText         = colorRGB(236, 239, 244) // #ECEFF4
	nordSubtext      = colorRGB(136, 150, 168) // darker #8898AC
	nordFocus        = colorRGB(191, 97, 106)  // #BF616A
	nordBreak        = colorRGB(163, 190, 140) // #A3BE8C
	nordHighlight    = colorRGB(136, 192, 208) // #88C0D0
	nordDanger       = colorRGB(191, 97, 106)
	nordDisabled     = colorRGB(100, 100, 170)
	nordDisabledText = colorRGB(100, 110, 120)
)

var (
	lightBackground   = colorRGB(236, 239, 244) // #ECEFF4
	lightPanel        = colorRGB(216, 222, 233) // #D8DEE9
	lightPanelAlt     = colorRGB(229, 233, 240)
	lightPanelMuted   = colorRGB(205, 208, 214)
	lightText         = colorRGB(46, 52, 64) // #2E3440
	lightSubtext      = colorRGB(70, 80, 90)
	lightFocus        = colorRGB(191, 97, 106)  // #BF616A
	lightBreak        = colorRGB(163, 190, 140) // #A3BE8C
	lightHighlight    = colorRGB(136, 192, 208) // #88C0D0
	lightDanger       = colorRGB(191, 97, 106)
	lightDisabled     = colorRGB(120, 125, 130)
	lightDisabledText = colorRGB(160, 165, 170)
)

var (
	appBackgroundColor      color.Color = nordBackground
	cardBackgroundColor     color.Color = nordPanel
	mutedTextColor          color.Color = nordSubtext
	workAccentColor         color.Color = nordFocus
	shortBreakAccentColor   color.Color = nordBreak
	longBreakAccentColor    color.Color = nordHighlight
	secondaryButtonColor    color.Color = nordPanelAlt
	toolbarCardColor        color.Color = nordPanel
	secondaryTextColor      color.Color = nordText
	primaryButtonColor      color.Color = nordHighlight
	primaryButtonTextColor  color.Color = nordText
	pauseButtonColor        color.Color = nordPanelMuted
	pauseButtonTextColor    color.Color = nordText
	disabledButtonColor     color.Color = nordPanelMuted
	disabledButtonTextColor color.Color = nordSubtext
	currentTheme            string      = "dark"
)

func CurrentThemeName() string {
	return currentTheme
}

func ApplyTheme(theme string) {
	currentTheme = theme
	if theme == "light" {
		appBackgroundColor = lightBackground
		cardBackgroundColor = lightPanel
		mutedTextColor = lightSubtext
		workAccentColor = lightFocus
		shortBreakAccentColor = lightBreak
		longBreakAccentColor = lightHighlight
		secondaryButtonColor = lightPanelAlt
		toolbarCardColor = lightPanel
		secondaryTextColor = lightText
		primaryButtonColor = lightHighlight
		primaryButtonTextColor = lightText
		pauseButtonColor = lightPanelMuted
		pauseButtonTextColor = lightText
		disabledButtonColor = lightDisabled
		disabledButtonTextColor = lightDisabledText
	} else {
		appBackgroundColor = nordBackground
		cardBackgroundColor = nordPanel
		mutedTextColor = nordSubtext
		workAccentColor = nordFocus
		shortBreakAccentColor = nordBreak
		longBreakAccentColor = nordHighlight
		secondaryButtonColor = nordPanelAlt
		toolbarCardColor = nordPanel
		secondaryTextColor = nordText
		primaryButtonColor = nordHighlight
		primaryButtonTextColor = nordText
		pauseButtonColor = nordPanelMuted
		pauseButtonTextColor = nordText
		disabledButtonColor = nordDisabled
		disabledButtonTextColor = nordDisabledText
	}
}

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
