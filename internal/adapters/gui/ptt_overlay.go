//go:build gui

// Draws the held push-to-talk control and routes its pointer capture.
// Speech sessions remain targeted to the originating editable field.
package gui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
)

func (game *Game) drawPTTButton(screen *ebiten.Image) {
	x, y := game.pttCenter()
	fill := color.RGBA{R: 38, G: 90, B: 156, A: 230}
	if game.pttHeld {
		fill = color.RGBA{R: 30, G: 150, B: 110, A: 240}
	}
	vector.DrawFilledCircle(screen, float32(x), float32(y), float32(pttRadius), fill, true)
	vector.StrokeCircle(screen, float32(x), float32(y), float32(pttRadius), 2, color.RGBA{R: 220, G: 232, B: 255, A: 255}, true)
	ebitenutil.DebugPrintAt(screen, "PTT", x-10, y-4)
	ebitenutil.DebugPrintAt(screen, game.voice.preview, windowMargin, y-24)
}

func (game *Game) pointInPTT(x int, y int) bool {
	if game.inWorkspace() && game.ui.GetFocusedWidget() == nil && game.voice.canvas == nil {
		return false
	}
	cx, cy := game.pttCenter()
	dx, dy := x-cx, y-cy
	return dx*dx+dy*dy <= pttRadius*pttRadius
}

func (game *Game) setPTTHeld(held bool) {
	if held == game.pttHeld {
		return
	}
	game.pttHeld = held
	if held {
		game.beginPTT("floating-ptt")
		return
	}
	game.endPTT()
}
