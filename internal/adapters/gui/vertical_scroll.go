//go:build gui

// Manages bounded vertical widget scrolling.
// Dependencies: github.com/ebitenui/ebitenui/widget, github.com/hajimehoshi/ebiten/v2, github.com/hajimehoshi/ebiten/v2/inpututil.
package gui

import (
	"image"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	bodyScrollDragThreshold      = 18
	bodyScrollCancelUpdateFrames = 2
)

func (game *Game) registerVerticalScroll(scroll *widget.ScrollContainer) {
	game.verticalScrolls = append(game.verticalScrolls, scroll)
}

func (game *Game) verticalScrollAt(x int, y int) *widget.ScrollContainer {
	point := imagePoint(x, y)
	for index := len(game.verticalScrolls) - 1; index >= 0; index-- {
		scroll := game.verticalScrolls[index]
		if point.In(scroll.ViewRect()) {
			return scroll
		}
	}
	return nil
}

func (game *Game) updateBodyScrollMouse(x int, y int) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && !game.pointInPTT(x, y) && !game.pointInDebugOverlay(x, y) {
		{
			scroll := game.verticalScrollAt(x, y)
			if scroll != nil && game.claimGesture(gestureBody, false, 0) {
				game.bodyScroll = scroll
				// One owner captures the complete press/drag/release sequence.
				game.bodyScrollLastY = y
				game.bodyScrollDragDistance = 0
				game.bodyScrollDragMoved = false
				game.lockBodyTabScroll()
			}
		}
	}
	if game.gesture.Owner == gestureBody && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		game.scrollBodyBy(y - game.bodyScrollLastY)
		game.bodyScrollLastY = y
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		game.finishBodyScroll()
	}
	{
		scroll := game.verticalScrollAt(x, y)
		if scroll != nil {
			_, wheelY := ebiten.Wheel()
			if wheelY != 0 {
				moveBodyScroll(scroll, wheelY*48)
			}
		}
	}
}

func (game *Game) beginBodyTouch(id ebiten.TouchID, x int, y int) bool {
	scroll := game.verticalScrollAt(x, y)
	if scroll == nil || !game.claimGesture(gestureBody, true, id) {
		return false
	}
	game.bodyScroll = scroll

	game.bodyScrollLastY = y
	game.bodyScrollDragDistance = 0
	game.bodyScrollDragMoved = false
	game.lockBodyTabScroll()
	return true
}

func (game *Game) updateBodyTouch() {
	if inpututil.IsTouchJustReleased(game.gesture.ID) || game.bodyScroll == nil {
		game.finishBodyScroll()
		game.releaseGesture(gestureBody)
		return
	}
	_, y := ebiten.TouchPosition(game.gesture.ID)
	game.scrollBodyBy(y - game.bodyScrollLastY)
	game.bodyScrollLastY = y
}

func (game *Game) scrollBodyBy(dy int) {
	if game.bodyScroll == nil {
		return
	}
	if dy < 0 {
		game.bodyScrollDragDistance -= dy
	} else {
		game.bodyScrollDragDistance += dy
	}
	if game.bodyScrollDragDistance >= bodyScrollDragThreshold {
		game.bodyScrollDragMoved = true
	}
	moveBodyScroll(game.bodyScroll, float64(dy))
}

func scrollFraction(current, delta float64, overflow int) float64 {
	if overflow <= 0 {
		return 0
	}
	return min(1, max(0, current-delta/float64(overflow)))
}
func moveBodyScroll(s *widget.ScrollContainer, delta float64) {
	if s.ViewRect().Dy() > 0 && s.ContentRect() != (image.Rectangle{}) {
		s.ScrollTop = scrollFraction(s.ScrollTop, delta, s.ContentRect().Dy()-s.ViewRect().Dy())
	}
}
func (game *Game) finishBodyScroll() {
	if game.bodyScrollDragMoved {
		game.bodyScrollCancelUpdates = bodyScrollCancelUpdateFrames
	} else {
		game.bodyTabScrollLocked = false
	}
	game.bodyScroll = nil
	game.releaseGesture(gestureBody)
	game.bodyScrollDragDistance = 0
	game.bodyScrollDragMoved = false
}

func (game *Game) bodySelectionSuppressed() bool {
	return game.gesture.Owner == gestureDebug || game.gesture.Owner == gesturePTT || game.bodyScrollDragMoved || game.bodyScrollCancelUpdates > 0
}

func (game *Game) advanceBodyScrollCancellation() {
	if game.bodyScrollCancelUpdates > 0 {
		game.bodyScrollCancelUpdates--
		if game.bodyScrollCancelUpdates == 0 {
			game.bodyTabScrollLocked = false
		}
	}
}

func (game *Game) lockBodyTabScroll() {
	if game.tabScroll == nil {
		return
	}
	game.bodyTabScrollLeft = game.tabScroll.ScrollLeft
	game.bodyTabScrollLocked = true
}

func (game *Game) enforceBodyTabScrollLock() {
	if game.bodyTabScrollLocked && game.tabScroll != nil {
		game.tabScroll.ScrollLeft = game.bodyTabScrollLeft
	}
}
