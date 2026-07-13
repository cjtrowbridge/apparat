//go:build gui

package gui

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const bodyScrollDragThreshold = 18

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
		if scroll := game.verticalScrollAt(x, y); scroll != nil {
			game.bodyScroll = scroll
			game.bodyScrollDragging = true
			game.bodyScrollLastY = y
			game.bodyScrollDragDistance = 0
			game.bodyScrollDragMoved = false
		}
	}
	if game.bodyScrollDragging && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		game.scrollBodyBy(y - game.bodyScrollLastY)
		game.bodyScrollLastY = y
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		game.finishBodyScroll()
	}
	if scroll := game.verticalScrollAt(x, y); scroll != nil {
		_, wheelY := ebiten.Wheel()
		if wheelY != 0 {
			scroll.ScrollTop -= wheelY / 12
		}
	}
}

func (game *Game) beginBodyTouch(id ebiten.TouchID, x int, y int) bool {
	scroll := game.verticalScrollAt(x, y)
	if scroll == nil {
		return false
	}
	game.bodyScroll = scroll
	game.bodyTouchActive = true
	game.bodyTouchID = id
	game.bodyScrollLastY = y
	game.bodyScrollDragDistance = 0
	game.bodyScrollDragMoved = false
	return true
}

func (game *Game) updateBodyTouch() {
	if inpututil.IsTouchJustReleased(game.bodyTouchID) || game.bodyScroll == nil {
		game.finishBodyScroll()
		game.bodyTouchActive = false
		return
	}
	_, y := ebiten.TouchPosition(game.bodyTouchID)
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
	game.bodyScroll.ScrollTop -= float64(dy) / 720
}

func (game *Game) finishBodyScroll() {
	if game.bodyScrollDragMoved {
		game.bodyScrollCancelUpdates = 2
	}
	game.bodyScroll = nil
	game.bodyScrollDragging = false
	game.bodyScrollDragDistance = 0
	game.bodyScrollDragMoved = false
}

func (game *Game) bodySelectionSuppressed() bool {
	return game.bodyScrollDragMoved || game.bodyScrollCancelUpdates > 0
}

func (game *Game) advanceBodyScrollCancellation() {
	if game.bodyScrollCancelUpdates > 0 {
		game.bodyScrollCancelUpdates--
	}
}
