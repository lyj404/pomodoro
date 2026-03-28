package ui

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed app-icon.png
var appIconBytes []byte

func AppIcon() fyne.Resource {
	return fyne.NewStaticResource("app-icon.png", appIconBytes)
}
