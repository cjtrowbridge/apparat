//go:build gui

package gui

import (
	"image/color"

	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

type tableRow []widget.PreferredSizeLocateableWidget

func (game *Game) borderedTable(header tableRow, rows ...tableRow) *widget.Container {
	columns := len(header)
	if columns == 0 {
		panic("table requires at least one header cell")
	}
	stretchedColumns := make([]bool, columns)
	for index := range stretchedColumns {
		stretchedColumns[index] = true
	}
	table := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(columns),
			widget.GridLayoutOpts.Spacing(2, 2),
			widget.GridLayoutOpts.Stretch(stretchedColumns, []bool{false}),
			widget.GridLayoutOpts.DefaultStretch(true, false),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{StretchHorizontal: true, StretchVertical: true})),
	)
	game.addTableRow(table, columns, header)
	for _, row := range rows {
		game.addTableRow(table, columns, row)
	}
	bordered := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(tableBorderImage(panelColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(widget.AnchorLayoutOpts.Padding(&widget.Insets{Left: 1, Right: 1, Top: 1, Bottom: 1}))),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})),
	)
	bordered.AddChild(table)
	return bordered
}

func (game *Game) addTableRow(table *widget.Container, columns int, row tableRow) {
	if len(row) != columns {
		panic("table row has inconsistent column count")
	}
	for _, value := range row {
		cell := widget.NewContainer(
			widget.ContainerOpts.BackgroundImage(tableBorderImage(panelColor)),
			widget.ContainerOpts.Layout(widget.NewAnchorLayout(widget.AnchorLayoutOpts.Padding(&widget.Insets{Left: 1, Right: 1, Top: 1, Bottom: 1}))),
			widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(0, 44)),
		)
		value.GetWidget().LayoutData = widget.AnchorLayoutData{StretchHorizontal: true, StretchVertical: true}
		cell.AddChild(value)
		table.AddChild(cell)
	}
}

func (game *Game) tableText(value string) *widget.Container {
	cell := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewAnchorLayout(widget.AnchorLayoutOpts.Padding(&widget.Insets{Left: 11, Right: 11, Top: 7, Bottom: 7}))))
	cell.AddChild(game.detailText(value))
	return cell
}

func tableBorderImage(fill color.Color) *uiimage.NineSlice {
	canvas := ebiten.NewImage(3, 3)
	canvas.Fill(color.RGBA{R: 88, G: 100, B: 126, A: 255})
	canvas.Set(1, 1, fill)
	return uiimage.NewNineSliceSimple(canvas, 1, 1)
}
