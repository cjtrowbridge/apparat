//go:build gui

// Places the chooser over its inactive originating page.
// A complete pointer gesture on the scrim dismisses without activating the page below.
package gui

import (
	"image"
	"image/color"

	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

func (g *Game) choiceBounds() image.Rectangle {
	x, y := g.width/6, g.height/6
	return image.Rect(x, y, g.width-x, g.height-y)
}

func (g *Game) choiceRoot(scroll *widget.ScrollContainer) *widget.Container {
	bounds := g.choiceBounds()
	outside := false
	root := widget.NewContainer(widget.ContainerOpts.BackgroundImage(uiimage.NewNineSliceColor(color.NRGBA{A: 150})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(widget.AnchorLayoutOpts.Padding(&widget.Insets{Left: bounds.Min.X, Top: bounds.Min.Y, Right: g.width - bounds.Max.X, Bottom: g.height - bounds.Max.Y}))),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MouseButtonPressedHandler(func(e *widget.WidgetMouseButtonPressedEventArgs) {
			outside = e.Button == ebiten.MouseButtonLeft && !image.Pt(e.OffsetX, e.OffsetY).In(bounds)
		}), widget.WidgetOpts.MouseButtonReleasedHandler(func(e *widget.WidgetMouseButtonReleasedEventArgs) {
			if outside && e.Button == ebiten.MouseButtonLeft && e.Inside && !image.Pt(e.OffsetX, e.OffsetY).In(bounds) && g.actions.mode == "choose" {
				g.cancelPTT()
				g.backActionScreen()
			}
			outside = false
		})))
	panel := widget.NewContainer(widget.ContainerOpts.BackgroundImage(g.theme.PanelTheme.BackgroundImage),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{StretchHorizontal: true, StretchVertical: true})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(widget.AnchorLayoutOpts.Padding(&widget.Insets{Top: 8, Left: 8, Right: 8, Bottom: pttRadius*2 + 8}))))
	panel.AddChild(scroll)
	root.AddChild(panel)
	return root
}
