//go:build gui

// Arbitrates one pointer owner across touch and mouse controls.
// Cancellation on geometry/focus changes never submits a voice capture.
package gui

import "github.com/hajimehoshi/ebiten/v2"

type gestureOwner uint8

const (
	gestureNone gestureOwner = iota
	gestureDebug
	gestureTabs
	gestureBody
	gestureDivider
	gesturePTT
)

type pointerGesture struct {
	Owner   gestureOwner
	Touch   bool
	ID      ebiten.TouchID
	Movable bool
}

func (game *Game) claimGesture(owner gestureOwner, touch bool, id ebiten.TouchID) bool {
	if game.gesture.Owner != gestureNone {
		return false
	}
	game.gesture = pointerGesture{Owner: owner, Touch: touch, ID: id, Movable: true}
	return true
}
func (game *Game) releaseGesture(owner gestureOwner) {
	if game.gesture.Owner == owner {
		if owner == gestureDebug || owner == gesturePTT {
			game.bodyScrollCancelUpdates = 2
			game.tabDragCancelUpdates = 2
		}
		game.gesture = pointerGesture{}
	}
}
func (game *Game) cancelPointer() {
	if game.gesture.Owner != gestureNone {
		game.bodyScrollCancelUpdates = 2
		game.tabDragCancelUpdates = 2
	}
	game.shell.CancelVoiceCapture()
	game.setPTTHeld(false)
	game.finishBodyScroll()
	game.finishTabStripDrag()
	game.gesture = pointerGesture{}
	game.rightCtrlHeld = false
	game.r2Held = false
}

// RequestInputCancellation crosses the Android UI thread without touching widgets.
func (game *Game) RequestInputCancellation() {
	game.cancelInputRequested.Store(true)
}
