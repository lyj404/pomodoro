package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type ActionTile struct {
	widget.BaseWidget
	label     string
	icon      fyne.Resource
	bgColor   color.Color
	textColor color.Color
	tapped    func()
	disabled  bool
}

func NewActionTile(label string, icon fyne.Resource, bgColor color.Color, textColor color.Color, tapped func()) *ActionTile {
	tile := &ActionTile{
		label:     label,
		icon:      icon,
		bgColor:   bgColor,
		textColor: textColor,
		tapped:    tapped,
	}
	tile.ExtendBaseWidget(tile)
	return tile
}

func (t *ActionTile) SetDisabled(disabled bool) {
	t.disabled = disabled
	t.Refresh()
}

func (t *ActionTile) RefreshColors(bgColor, textColor color.Color) {
	t.bgColor = bgColor
	t.textColor = textColor
	t.Refresh()
}

func (t *ActionTile) Tapped(*fyne.PointEvent) {
	if t.disabled || t.tapped == nil {
		return
	}
	t.tapped()
}

func (t *ActionTile) TappedSecondary(*fyne.PointEvent) {}

func (t *ActionTile) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(t.bgColor)
	bg.CornerRadius = 18

	icon := widget.NewIcon(t.icon)
	if t.icon == nil {
		icon.Hide()
	}
	label := canvas.NewText(t.label, t.textColor)
	label.TextSize = 14
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.Alignment = fyne.TextAlignCenter

	body := container.NewHBox(
		layout.NewSpacer(),
		icon,
		label,
		layout.NewSpacer(),
	)

	root := container.NewStack(bg, container.NewPadded(body))
	return &actionTileRenderer{
		tile:    t,
		bg:      bg,
		icon:    icon,
		label:   label,
		objects: []fyne.CanvasObject{root},
	}
}

type actionTileRenderer struct {
	tile    *ActionTile
	bg      *canvas.Rectangle
	icon    *widget.Icon
	label   *canvas.Text
	objects []fyne.CanvasObject
}

func (r *actionTileRenderer) Layout(size fyne.Size) {
	r.objects[0].Resize(size)
}

func (r *actionTileRenderer) MinSize() fyne.Size {
	return fyne.NewSize(148, 46)
}

func (r *actionTileRenderer) Refresh() {
	bg := r.tile.bgColor
	text := r.tile.textColor
	if r.tile.disabled {
		bg = disabledButtonColor
		text = disabledButtonTextColor
	}
	r.bg.FillColor = bg
	r.label.Color = text
	r.icon.SetResource(r.tile.icon)
	r.label.Text = r.tile.label
	r.bg.Refresh()
	r.label.Refresh()
	r.icon.Refresh()
}

func (r *actionTileRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *actionTileRenderer) Destroy() {}

func (r *actionTileRenderer) BackgroundColor() color.Color {
	return color.Transparent
}
