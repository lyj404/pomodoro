package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type FocusTheme struct{}

type LightTheme struct{}

func NewFocusTheme() fyne.Theme {
	return &FocusTheme{}
}

func NewLightTheme() fyne.Theme {
	return &LightTheme{}
}

func (t *FocusTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return nordBackground
	case theme.ColorNameButton:
		return nordPanelAlt
	case theme.ColorNameDisabledButton:
		return nordPanelMuted
	case theme.ColorNameDisabled:
		return nordPanelMuted
	case theme.ColorNameForeground:
		return nordText
	case theme.ColorNamePrimary:
		return nordHighlight
	case theme.ColorNameInputBackground:
		return nordPanel
	case theme.ColorNamePlaceHolder:
		return nordSubtext
	case theme.ColorNameHover:
		return nordPanelMuted
	case theme.ColorNameFocus:
		return nordHighlight
	case theme.ColorNameSelection:
		return nordPanelAlt
	case theme.ColorNameScrollBar:
		return nordPanelMuted
	case theme.ColorNameShadow:
		return nordBackground
	case theme.ColorNameOverlayBackground:
		return &color.NRGBA{R: 59, G: 66, B: 82, A: 245}
	case theme.ColorNameMenuBackground:
		return nordPanel
	case theme.ColorNameInputBorder:
		return nordText
	case theme.ColorNameSeparator:
		return nordSubtext
	case theme.ColorNamePressed:
		return nordHighlight
	case theme.ColorNameHyperlink:
		return nordHighlight
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (t *LightTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return lightBackground
	case theme.ColorNameButton:
		return lightPanelAlt
	case theme.ColorNameDisabledButton:
		return lightPanelMuted
	case theme.ColorNameDisabled:
		return lightPanelMuted
	case theme.ColorNameForeground:
		return lightText
	case theme.ColorNamePrimary:
		return lightHighlight
	case theme.ColorNameInputBackground:
		return lightPanel
	case theme.ColorNamePlaceHolder:
		return lightSubtext
	case theme.ColorNameHover:
		return lightPanelMuted
	case theme.ColorNameFocus:
		return lightHighlight
	case theme.ColorNameSelection:
		return lightPanelAlt
	case theme.ColorNameScrollBar:
		return lightPanelMuted
	case theme.ColorNameShadow:
		return lightBackground
	case theme.ColorNameOverlayBackground:
		return &color.NRGBA{R: 216, G: 222, B: 233, A: 245}
	case theme.ColorNameMenuBackground:
		return lightPanel
	case theme.ColorNameInputBorder:
		return lightText
	case theme.ColorNameSeparator:
		return lightSubtext
	case theme.ColorNamePressed:
		return lightHighlight
	case theme.ColorNameHyperlink:
		return lightHighlight
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (t *FocusTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *FocusTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *FocusTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameInnerPadding:
		return 10
	case theme.SizeNamePadding:
		return 14
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameInputBorder:
		return 1.5
	case theme.SizeNameText:
		return 15
	}
	return theme.DefaultTheme().Size(name)
}

func (t *LightTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *LightTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *LightTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameInnerPadding:
		return 10
	case theme.SizeNamePadding:
		return 14
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameInputBorder:
		return 1.5
	case theme.SizeNameText:
		return 15
	}
	return theme.DefaultTheme().Size(name)
}
