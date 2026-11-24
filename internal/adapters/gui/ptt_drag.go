//go:build gui

// Owns pointer capture before the active HUD or game surface receives input.
// Intersecting equal-radius targets share gesture transitions on mouse and touch.
package gui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
)

const pttTravel = 3 * pttRadius / 2

func (g *Game) pttDisplayCenter() (int, int) {
	x, y := g.pttCenter()
	if g.voice.tap.slideLocked {
		return x, y - pttTravel
	}
	if g.gesture.global {
		return x - pttTravel, y
	}
	return x, y
}

func (g *Game) drawPTTTargets(screen *ebiten.Image) {
	if g.gesture.Owner != gesturePTT && !g.voice.tap.slideLocked {
		return
	}
	x, y := g.pttCenter()
	target := func(cx, cy int, label string) {
		vector.DrawFilledCircle(screen, float32(cx), float32(cy), pttRadius, color.RGBA{42, 53, 75, 225}, true)
		vector.StrokeCircle(screen, float32(cx), float32(cy), pttRadius, 2, color.White, true)
		ebitenutil.DebugPrintAt(screen, label, cx-len(label)*3, cy-8)
	}
	target(x, y, "PTT")
	target(x, y-pttTravel, "Lock")
	target(x-pttTravel, y, "Agent")
}

func (g *Game) pressPTTPointer(x, y int, touch bool, id ebiten.TouchID) bool {
	if !g.pointInPTT(x, y) || !g.claimGesture(gesturePTT, touch, id) {
		return false
	}
	g.gesture.startX, g.gesture.startY = x, y
	if g.voice.tap.slideLocked {
		g.pttHeld = true
	} else {
		g.setPTTHeld(true)
	}
	return true
}

func (g *Game) dragPTTPointer(x, y int) {
	if g.gesture.global {
		return
	}
	dx, dy := x-g.gesture.startX, y-g.gesture.startY
	near := func(x, y int) bool {
		return x*x+y*y <= pttRadius*pttRadius/2
	}
	lock := &g.voice.tap
	if lock.slideLocked {
		if near(dx, dy-pttTravel) {
			*lock = pttLock{}
			g.endPTT()
			g.pttHeld = false
		}
		return
	}
	if !g.pttHeld {
		return
	}
	if near(dx+pttTravel, dy) {
		g.gesture.global = true
		g.openGlobalPrompt()
	} else if near(dx, dy+pttTravel) && lock.source == "floating-ptt" {
		lock.locked = true
		lock.slideLocked = true
		g.gesture.startX, g.gesture.startY = x, y
	}
}

func (g *Game) updatePTTPointer() bool {
	if !ebiten.IsFocused() {
		g.cancelPointer()
		return false
	}
	if g.gesture.Owner == gestureNone {
		for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
			x, y := ebiten.TouchPosition(id)
			if g.pressPTTPointer(x, y, true, id) {
				return true
			}
		}
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			return g.pressPTTPointer(x, y, false, 0)
		}
	}
	if g.gesture.Owner != gesturePTT {
		return false
	}
	x, y := ebiten.CursorPosition()
	released := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	if g.gesture.Touch {
		released = inpututil.IsTouchJustReleased(g.gesture.ID)
		if released {
			x, y = inpututil.TouchPositionInPreviousTick(g.gesture.ID)
		} else {
			x, y = ebiten.TouchPosition(g.gesture.ID)
		}
	}
	g.dragPTTPointer(x, y)
	if released {
		if g.gesture.global && g.voice.tap.source == "floating-ptt" {
			g.voice.tap = pttLock{}
			g.endPTT()
		}
		g.setPTTHeld(false)
		g.releaseGesture(gesturePTT)
	}
	return true
}
