//go:build gui

// Draws the held push-to-talk control and routes its pointer capture.
// Fill and ring distinguish held and locked capture with a stable PTT label.
package gui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
)

func (game *Game) drawPTTButton(screen *ebiten.Image) {
	x, y := game.pttDisplayCenter()
	game.drawPTTTargets(screen)
	fill := game.pttButtonFill()
	vector.DrawFilledCircle(screen, float32(x), float32(y), float32(pttRadius), fill, true)
	vector.StrokeCircle(screen, float32(x), float32(y), float32(pttRadius), 2, color.RGBA{R: 220, G: 232, B: 255, A: 255}, true)
	ebitenutil.DebugPrintAt(screen, "PTT", x-9, y-8)
	if game.voice.tap.locked {
		vector.StrokeCircle(screen, float32(x), float32(y), float32(pttRadius-5), 2, color.White, true)
	}
	ebitenutil.DebugPrintAt(screen, game.voice.preview, windowMargin, y-24)
}

func (game *Game) pointInPTT(x int, y int) bool {
	cx, cy := game.pttDisplayCenter()
	dx, dy := x-cx, y-cy
	return dx*dx+dy*dy <= pttRadius*pttRadius
}

func (game *Game) setPTTHeld(held bool) {
	if held == game.pttHeld {
		return
	}
	game.pttHeld = held
	game.tapPTT("floating-ptt", held)
}

const pttTriggerButton = ebiten.StandardGamepadButtonFrontBottomLeft

func (game *Game) pttButtonHeld() bool {
	return game.voice.tap.locked || game.pttHeld || game.l2Held || game.rightCtrlHeld
}

func (game *Game) pttButtonFill() color.RGBA {
	if game.pttButtonHeld() {
		return color.RGBA{R: 30, G: 150, B: 110, A: 240}
	}
	return color.RGBA{R: 38, G: 90, B: 156, A: 230}
}

func (game *Game) setTriggerPTT(held bool) {
	if held == game.l2Held {
		return
	}
	game.l2Held = held
	game.tapPTT("gamepad-l2", held)
}

func (game *Game) cancelTriggerPTT() {
	if game.l2Held || game.voice.tap.source == "gamepad-l2" {
		game.l2Held = false
		game.cancelPTT()
	}
}

func (game *Game) keyboardPTT() {
	rightCtrl := ebiten.IsKeyPressed(ebiten.KeyControlRight)
	if rightCtrl && !game.rightCtrlHeld {
		game.beginPTT("right-ctrl")
	}
	if !rightCtrl && game.rightCtrlHeld && game.voice.tap.source == "" {
		game.endPTT()
	}
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		game.cancelPTT()
		game.setPTTHeld(false)
	}
	game.rightCtrlHeld = rightCtrl
}

func (game *Game) arcadeTriggerPTT() {
	for _, id := range ebiten.AppendGamepadIDs(nil) {
		if ebiten.IsStandardGamepadLayoutAvailable(id) {
			game.setTriggerPTT(ebiten.IsStandardGamepadButtonPressed(id, pttTriggerButton))
			return
		}
	}
	game.cancelTriggerPTT()
}
