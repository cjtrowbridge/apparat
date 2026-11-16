//go:build gui

// Tests exclusive gesture capture and cancellation across pointer sources.
// No synthetic OS input is needed for these ownership transitions.
package gui

import (
	"github.com/cjtrowbridge/apparat/internal/hud"
	"testing"
)

func TestGestureOwnershipAndCancellation(t *testing.T) {
	game := NewGame()
	if !game.claimGesture(gesturePTT, true, 7) {
		t.Fatal("first pointer rejected")
	}
	game.setPTTHeld(true)
	if game.claimGesture(gestureTabs, true, 8) || game.claimGesture(gestureBody, false, 0) {
		t.Fatal("competing pointer stole capture")
	}
	game.releaseGesture(gestureTabs)
	if game.gesture.Owner != gesturePTT {
		t.Fatal("wrong owner released capture")
	}
	game.cancelPointer()
	if game.gesture.Owner != gestureNone || game.pttHeld || game.shell.Snapshot().VoiceState != hud.VoiceIdle {
		t.Fatal("cancellation submitted capture")
	}
	if !game.claimGesture(gestureBody, false, 0) {
		t.Fatal("capture not reusable")
	}
	game.Layout(360, 800)
	if game.gesture.Owner != gestureNone {
		t.Fatal("rotation retained gesture")
	}
}

func TestOverlayAndPTTCaptureSuppressUnderlyingActivation(t *testing.T) {
	for _, owner := range []gestureOwner{gestureDebug, gesturePTT} {
		game := NewGame()
		game.claimGesture(owner, true, 1)
		if !game.bodySelectionSuppressed() || !game.tabSelectionSuppressed() {
			t.Fatal("underlying controls admitted during capture")
		}
		game.releaseGesture(owner)
		if !game.bodySelectionSuppressed() || !game.tabSelectionSuppressed() {
			t.Fatal("deferred release escaped capture")
		}
		for range 2 {
			game.advanceBodyScrollCancellation()
			game.advanceTabDragCancellation()
		}
		if game.bodySelectionSuppressed() || game.tabSelectionSuppressed() {
			t.Fatal("capture suppression stuck")
		}
	}
}
