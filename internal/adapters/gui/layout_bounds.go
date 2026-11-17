//go:build gui

// Computes bounded tab and content geometry.
// Dependencies: github.com/ebitenui/ebitenui/input, github.com/ebitenui/ebitenui/widget, github.com/hajimehoshi/ebiten/v2.
package gui

import (
	"image"
	"image/color"
	"strings"

	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	ebitentext "github.com/hajimehoshi/ebiten/v2/text/v2"
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
	{
		relayoutable, ok := b.child.(widget.Relayoutable)
		if ok {
			relayoutable.RequestRelayout()
		}
	}
}

func (b *boundedPreferredWidget) Render(screen *ebiten.Image) {
	{
		renderer, ok := b.child.(widget.Renderer)
		if ok {
			renderer.Render(screen)
		}
	}
}

func (b *boundedPreferredWidget) Update(updObj *widget.UpdateObject) {
	{
		updater, ok := b.child.(widget.Updater)
		if ok {
			updater.Update(updObj)
		}
	}
}

func (b *boundedPreferredWidget) SetupInputLayer(def input.DeferredSetupInputLayerFunc) {
	{
		layerer, ok := b.child.(input.Layerer)
		if ok {
			layerer.SetupInputLayer(def)
		}
	}
}

func (b *boundedPreferredWidget) GetFocusers() []widget.Focuser {
	{
		container, ok := b.child.(interface{ GetFocusers() []widget.Focuser })
		if ok {
			return container.GetFocusers()
		}
	}
	return nil
}

func (b *boundedPreferredWidget) GetDropTargets() []widget.HasWidget {
	{
		dropper, ok := b.child.(widget.Dropper)
		if ok {
			return dropper.GetDropTargets()
		}
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
	// Keep readable detail text outside the floating PTT capture region.
	width := game.hudPreferredWidth() - 72 - (2*pttRadius + pttMargin)
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

// Keep words intact where possible, then break long names at rune boundaries.
func wrapSelectorLabel(label string, face ebitentext.Face, width int) string {
	fits := func(line string) bool {
		measured, _ := ebitentext.Measure(line, face, 0)
		return measured <= float64(width)
	}
	if fits(label) {
		return label
	}

	var lines []string
	line := ""
	for _, word := range strings.Fields(label) {
		if line != "" {
			candidate := line + " " + word
			if fits(candidate) {
				line = candidate
				continue
			}
			lines = append(lines, line)
			line = ""
		}
		if fits(word) {
			line = word
			continue
		}
		for _, letter := range word {
			candidate := line + string(letter)
			if line != "" && !fits(candidate) {
				lines = append(lines, line)
				line = string(letter)
			} else {
				line = candidate
			}
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
