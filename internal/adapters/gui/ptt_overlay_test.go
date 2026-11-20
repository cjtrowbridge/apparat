//go:build gui

// Verifies identical visible feedback across PTT sources and cancellation on controller loss.
// Uses the renderer's actual fill selection without opening a microphone session.
package gui

import (
	"testing"

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
