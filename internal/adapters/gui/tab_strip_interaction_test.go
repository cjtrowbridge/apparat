//go:build gui

package gui

import (
	"testing"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/widget"
)

func TestTabStripDragStateChangeKeepsOnlySelectedButtonChecked(t *testing.T) {
	game := NewGame()
	event.ExecuteDeferred()
	if len(game.tabButtons) < 3 {
		t.Fatal("tab strip did not retain enough buttons")
	}
	game.tabStripDragMoved = true
	game.tabButtons[2].SetState(widget.WidgetChecked)
	event.ExecuteDeferred()
	checked := 0
	for index, button := range game.tabButtons {
		if button.State() == widget.WidgetChecked {
			checked++
			if index != game.shell.Snapshot().ActiveIndex {
				t.Fatalf("button %d checked after drag, want only active %d", index, game.shell.Snapshot().ActiveIndex)
			}
		}
	}
	if checked != 1 {
		t.Fatalf("checked tab buttons after drag = %d, want 1", checked)
	}
}

func TestTabStripDragScrollsWithoutSelectingReleaseTarget(t *testing.T) {
	game := NewGame()
	game.width = 360
	game.rebuildUI(game.shell.Snapshot())
	event.ExecuteDeferred()
	game.tabScroll.ScrollLeft = 0.5
	game.scrollTabStripBy(-tabDragThreshold)
	if got := game.tabScroll.ScrollLeft; got <= 0.5 {
		t.Fatalf("tab strip scroll = %.2f, want greater than 0.50 after leftward drag", got)
	}
	game.finishTabStripDrag()
	game.tabButtons[2].SetState(widget.WidgetChecked)
	event.ExecuteDeferred()
	if active := game.shell.Snapshot().ActiveIndex; active != 0 {
		t.Fatalf("active tab after drag release = %d, want 0", active)
	}
	assertOnlyActiveTabChecked(t, game)
}

func TestTabStripDragReleaseKeepsSelectionSuppressedThroughDeferredEvents(t *testing.T) {
	game := NewGame()
	game.tabStripDragMoved = true
	if !game.tabSelectionSuppressed() {
		t.Fatal("tab selection was not suppressed during drag")
	}
	game.finishTabStripDrag()
	if game.tabStripDragMoved {
		t.Fatal("tab drag moved flag stayed set after release")
	}
	if !game.tabSelectionSuppressed() {
		t.Fatal("tab selection suppression did not survive release")
	}
	game.advanceTabDragCancellation()
	if !game.tabSelectionSuppressed() {
		t.Fatal("tab selection suppression ended before deferred events could run")
	}
	game.advanceTabDragCancellation()
	if game.tabSelectionSuppressed() {
		t.Fatal("tab selection suppression did not end after deferred event window")
	}
}

func TestTabRadioGroupTapSelectsExactlyOneTab(t *testing.T) {
	game := NewGame()
	event.ExecuteDeferred()
	if len(game.tabButtons) < 2 {
		t.Fatal("tab strip did not retain buttons")
	}
	game.tabRadioGroup.SetActive(game.tabButtons[1])
	event.ExecuteDeferred()
	if active := game.shell.Snapshot().ActiveIndex; active != 1 {
		t.Fatalf("active tab after radio selection = %d, want 1", active)
	}
	assertOnlyActiveTabChecked(t, game)
}

func TestTabStripRepeatedDragsAndKeyboardSelectionKeepOneCheckedTab(t *testing.T) {
	game := NewGame()
	event.ExecuteDeferred()
	for _, target := range []int{1, 2, 4} {
		game.tabStripDragMoved = true
		game.finishTabStripDrag()
		game.tabButtons[target].SetState(widget.WidgetChecked)
		event.ExecuteDeferred()
		assertOnlyActiveTabChecked(t, game)
		game.advanceTabDragCancellation()
		game.advanceTabDragCancellation()
	}
	if err := game.shell.ApplyAction(hud.ActionNextTab); err != nil {
		t.Fatal(err)
	}
	game.rebuildUI(game.shell.Snapshot())
	assertOnlyActiveTabChecked(t, game)
}

func assertOnlyActiveTabChecked(t *testing.T, game *Game) {
	t.Helper()
	active := game.shell.Snapshot().ActiveIndex
	checked := 0
	for index, button := range game.tabButtons {
		if button.State() != widget.WidgetChecked {
			continue
		}
		checked++
		if index != active {
			t.Fatalf("button %d checked, want only active %d", index, active)
		}
	}
	if checked != 1 {
		t.Fatalf("checked tab buttons = %d, want 1", checked)
	}
}
