package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

func dialogHeader(titleText string, subtitleText string, accent color.Color) fyne.CanvasObject {
	bar := canvas.NewRectangle(accent)
	bar.SetMinSize(fyne.NewSize(8, 40))

	title := canvas.NewText(titleText, nordText)
	title.TextSize = 18
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := canvas.NewText(subtitleText, nordSubtext)
	subtitle.TextSize = 13

	body := container.NewVBox(title, subtitle)
	card := canvas.NewRectangle(nordPanel)
	card.CornerRadius = 20

	return container.NewStack(
		card,
		container.NewBorder(nil, nil, bar, nil, container.NewPadded(body)),
	)
}

func mainHeader(titleText string, subtitleText string, trailing fyne.CanvasObject, accent color.Color) fyne.CanvasObject {
	bar := canvas.NewRectangle(accent)
	bar.SetMinSize(fyne.NewSize(10, 72))

	title := canvas.NewText(titleText, nordText)
	title.TextSize = 26
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := canvas.NewText(subtitleText, nordSubtext)
	subtitle.TextSize = 13

	body := container.NewBorder(nil, nil, nil, trailing, container.NewVBox(title, subtitle))
	card := canvas.NewRectangle(nordPanel)
	card.CornerRadius = 22

	return container.NewStack(
		card,
		container.NewBorder(nil, nil, bar, nil, container.NewPadded(body)),
	)
}

func centeredDialogContent(maxWidth float32, children ...fyne.CanvasObject) fyne.CanvasObject {
	sidePadding := canvas.NewRectangle(colorTransparent())
	sidePadding.SetMinSize(fyne.NewSize(20, 1))

	column := container.NewVBox(children...)
	innerRow := container.NewHBox(sidePadding, column, sidePadding)

	spacer := layout.NewSpacer()
	wrapWidth := maxWidth
	if wrapWidth > 500 {
		wrapWidth = 500
	}
	columnWrap := container.NewGridWrap(fyne.NewSize(wrapWidth, column.MinSize().Height), innerRow)
	return container.NewHBox(spacer, columnWrap, spacer)
}
