//go:build gui

package gui

import (
	"image"
	"image/color"

	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

type boundedPreferredWidget struct {
	child    widget.PreferredSizeLocateableWidget
	maxWidth func() int
}

func boundPreferredWidth(child widget.PreferredSizeLocateableWidget, maxWidth func() int) widget.PreferredSizeLocateableWidget {
	return &boundedPreferredWidget{child: child, maxWidth: maxWidth}
}

func (b *boundedPreferredWidget) GetWidget() *widget.Widget {
	return b.child.GetWidget()
}

func (b *boundedPreferredWidget) PreferredSize() (int, int) {
	width, height := b.child.PreferredSize()
	if b.maxWidth == nil {
		return width, height
	}
	limit := b.maxWidth()
	if limit > 0 && width > limit {
		width = limit
	}
	return width, height
}

func (b *boundedPreferredWidget) SetLocation(rect image.Rectangle) {
	b.child.SetLocation(rect)
}

func (b *boundedPreferredWidget) Validate() {
	b.child.Validate()
}

func (b *boundedPreferredWidget) RequestRelayout() {
	if relayoutable, ok := b.child.(widget.Relayoutable); ok {
		relayoutable.RequestRelayout()
	}
}

func (b *boundedPreferredWidget) Render(screen *ebiten.Image) {
	if renderer, ok := b.child.(widget.Renderer); ok {
		renderer.Render(screen)
	}
}

func (b *boundedPreferredWidget) Update(updObj *widget.UpdateObject) {
	if updater, ok := b.child.(widget.Updater); ok {
		updater.Update(updObj)
	}
}

func (b *boundedPreferredWidget) SetupInputLayer(def input.DeferredSetupInputLayerFunc) {
	if layerer, ok := b.child.(input.Layerer); ok {
		layerer.SetupInputLayer(def)
	}
}

func (b *boundedPreferredWidget) GetFocusers() []widget.Focuser {
	if container, ok := b.child.(interface{ GetFocusers() []widget.Focuser }); ok {
		return container.GetFocusers()
	}
	return nil
}

func (b *boundedPreferredWidget) GetDropTargets() []widget.HasWidget {
	if dropper, ok := b.child.(widget.Dropper); ok {
		return dropper.GetDropTargets()
	}
	return nil
}

func (game *Game) hudPreferredWidth() int {
	return max(120, game.width-(windowMargin*2))
}

func (game *Game) hudTextMaxWidth() float64 {
	return float64(max(120, game.hudPreferredWidth()-72))
}

func (game *Game) detailTextMaxWidth() float64 {
	width := game.hudPreferredWidth() - 72
	if !game.collapsed() {
		width -= game.masterListWidth() + 40
	}
	return float64(max(120, width))
}

func (game *Game) hudText(value string) *widget.Text {
	return widget.NewText(
		widget.TextOpts.Text(value, game.theme.ButtonTheme.TextFace, color.White),
		widget.TextOpts.MaxWidth(game.hudTextMaxWidth()),
	)
}

func (game *Game) detailText(value string) *widget.Text {
	return widget.NewText(
		widget.TextOpts.Text(value, game.theme.ButtonTheme.TextFace, color.White),
		widget.TextOpts.MaxWidth(game.detailTextMaxWidth()),
	)
}
