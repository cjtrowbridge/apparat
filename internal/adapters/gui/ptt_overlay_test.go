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
	if !game.pttButtonHeld() || game.pttButtonFill() != pressed {
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

func TestSlideLockSurvivesReleaseUntilDraggedHome(t *testing.T) {
	g := NewGame()
	defer g.Close()
	cx, cy := g.pttCenter()
	g.pressPTTPointer(cx, cy, true, 2)
	g.voice.tap = pttLock{source: "floating-ptt"}
	g.shell.StartVoiceCapture("floating-ptt")
	g.dragPTTPointer(cx, cy-pttTravel)
	g.setPTTHeld(false)
	g.releaseGesture(gesturePTT)
	if !g.voice.tap.slideLocked || !g.pointInPTT(cx, cy-pttTravel) || g.pointInPTT(cx, cy) {
		t.Fatal("locked circle must stay at the upper target")
	}
	g.setTriggerPTT(true)
	g.setTriggerPTT(false)
	g.pressPTTPointer(cx, cy-pttTravel, false, 0)
	if g.shell.Snapshot().VoiceState != hud.VoiceRecording {
		t.Fatal("grabbing the locked circle finished capture")
	}
	g.dragPTTPointer(cx, cy)
	if g.shell.Snapshot().VoiceState != hud.VoiceQueued || g.voice.tap.locked {
		t.Fatal("returning home did not finish capture")
	}
}

func TestDragIntoGlobalDraft(t *testing.T) {
	g := NewGame()
	defer g.Close()
	for _, width := range []int{320, 1100} {
		g.Layout(width, 740)
		g.rebuildUI(g.shell.Snapshot())
		g.openActionBrowser()
		x, y := g.pttCenter()
		g.pressPTTPointer(x, y, false, 0)
		g.dragPTTPointer(x-pttTravel, y)
		if g.actions.mode != "" || g.shell.Snapshot().ActiveTab().ID() != hud.TabGlobalSession || g.ui.GetFocusedWidget() != g.globalField {
			t.Fatal("drag did not enter the global editor")
		}
		g.setPTTHeld(false)
		g.releaseGesture(gesturePTT)
		renderNavigation(g)
		field := g.globalField.GetWidget().Rect
		cx, cy := g.pttCenter()
		if field.Dy() < 44 || field.Max.X > g.width || field.Min.X < 0 || field.Max.Y > cy-pttRadius-pttTravel && field.Max.X > cx-pttRadius {
			t.Fatalf("prompt outside its reserved space: %v at %d", field, width)
		}
		g.globalField.SetText("")
		g.finishUtterance(g.globalField, "open Projects")
		g.selectTab(0, nil)
		g.openGlobalPrompt()
		if g.globalField.GetText() != "open Projects" {
			t.Fatal("command-shaped draft lost on return")
		}
	}
}
