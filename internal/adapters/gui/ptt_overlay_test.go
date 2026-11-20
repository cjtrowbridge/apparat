//go:build gui

// Verifies identical visible feedback across PTT sources and cancellation on controller loss.
// Uses the renderer's actual fill selection without opening a microphone session.
package gui

import (
	"testing"
	"time"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestTriggerPTTMatchesPointerFeedbackAndRelease(t *testing.T) {
	game := NewGame()
	idle := game.pttButtonFill()
	game.setPTTHeld(true)
	pressed := game.pttButtonFill()
	game.setPTTHeld(false)
	if pressed == idle {
		t.Fatal("pointer PTT has no pressed feedback")
	}
	game.setTriggerPTT(true)
	if !game.pttButtonHeld() || !game.pttButtonVisible() || game.pttButtonFill() != pressed {
		t.Fatal("left trigger does not show the same pressed PTT appearance")
	}
	game.setTriggerPTT(false)
	if game.pttButtonHeld() || game.pttButtonFill() != idle {
		t.Fatal("PTT appearance did not reset on trigger release")
	}
	game.rightCtrlHeld = true
	if game.pttButtonFill() != pressed {
		t.Fatal("keyboard PTT does not share pressed feedback")
	}
	game.setTriggerPTT(true)
	game.setTriggerPTT(false)
	if game.pttButtonFill() != pressed {
		t.Fatal("trigger release cleared another held source's appearance")
	}
	if pttTriggerButton != ebiten.StandardGamepadButtonFrontBottomLeft {
		t.Fatal("PTT must use L2, not R2")
	}
	bindings := hud.DefaultBindings().Inputs(hud.ActionPushToTalk)
	if len(bindings) != 2 || bindings[0].Input != "L2" || bindings[1].Input != "RightCtrl" {
		t.Fatalf("PTT contract mismatch: %#v", bindings)
	}
}

func TestControllerLossCancelsPTTWithoutSubmission(t *testing.T) {
	game := NewGame()
	idle := game.pttButtonFill()
	game.setTriggerPTT(true)
	game.shell.StartVoiceCapture("gamepad-l2")
	game.cancelTriggerPTT()
	if game.l2Held || game.pttButtonFill() != idle || game.shell.Snapshot().VoiceState != hud.VoiceIdle {
		t.Fatal("controller loss must clear feedback and cancel rather than submit")
	}
}

func TestPTTLockTransitions(t *testing.T) {
	start := time.Unix(1, 0)
	for _, source := range []string{"floating-ptt", "gamepad-l2"} {
		var lock pttLock
		if lock.press(source, start) != tapBegin || lock.release(source, start.Add(60*time.Millisecond)) != tapWait {
			t.Fatal("first tap submitted early")
		}
		if lock.press(source, start.Add(160*time.Millisecond)) != tapWait || !lock.locked {
			t.Fatal("second press did not lock original session")
		}
		if lock.release(source, start.Add(200*time.Millisecond)) != tapWait || lock.expired(start.Add(time.Hour)) {
			t.Fatal("lock lost after release")
		}
		if lock.press("floating-ptt", start.Add(time.Hour)) != tapSubmit || lock.locked || lock.release("floating-ptt", start.Add(time.Hour)) != tapWait {
			t.Fatal("stop must submit exactly once")
		}
		lock.press(source, start)
		if lock.release(source, start.Add(time.Second)) != tapSubmit {
			t.Fatal("hold did not submit on release")
		}
		lock.press(source, start)
		lock.release(source, start)
		if lock.expired(start.Add(time.Millisecond)) || !lock.expired(start.Add(pttTapWindow)) || lock.expired(start.Add(time.Second)) {
			t.Fatal("single tap timeout must submit once")
		}
	}
}
