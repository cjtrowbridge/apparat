//go:build gui

// Draws the held push-to-talk control and routes its pointer capture.
// The displayed role follows the editor or voice command target captured on press.
package gui

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
)

func (game *Game) drawPTTButton(screen *ebiten.Image) {
	x, y := game.pttCenter()
	fill := game.pttButtonFill()
	vector.DrawFilledCircle(screen, float32(x), float32(y), float32(pttRadius), fill, true)
	vector.StrokeCircle(screen, float32(x), float32(y), float32(pttRadius), 2, color.RGBA{R: 220, G: 232, B: 255, A: 255}, true)
	label := "Voice\ncommand"
	_, editing := game.ui.GetFocusedWidget().(*widget.TextInput)
	if game.voice.canvas != nil {
		editing = game.voice.field != nil
	}
	if editing {
		label = "Dictate"
	}
	ebitenutil.DebugPrintAt(screen, label, x-21, y-10)
	ebitenutil.DebugPrintAt(screen, game.voice.preview, windowMargin, y-24)
}

func (game *Game) pointInPTT(x int, y int) bool {
	if !game.pttButtonVisible() {
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

const pttTriggerButton = ebiten.StandardGamepadButtonFrontBottomLeft

func (game *Game) pttButtonHeld() bool {
	return game.pttHeld || game.l2Held || game.rightCtrlHeld
}

func (game *Game) pttButtonVisible() bool {
	return game.actions.mode != "" || game.voice.engine != nil || !game.inWorkspace() || game.ui.GetFocusedWidget() != nil || game.voice.canvas != nil || game.pttButtonHeld()
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
	if held {
		game.beginPTT("gamepad-l2")
		return
	}
	game.endPTT()
}

func (game *Game) cancelTriggerPTT() {
	if game.l2Held {
		game.l2Held = false
		game.cancelPTT()
	}
}

func (game *Game) keyboardPTT() {
	rightCtrl := ebiten.IsKeyPressed(ebiten.KeyControlRight)
	if rightCtrl && !game.rightCtrlHeld {
		game.beginPTT("right-ctrl")
	}
	if !rightCtrl && game.rightCtrlHeld {
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
