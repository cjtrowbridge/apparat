//go:build gui

// Verifies manages bounded vertical widget scrolling.
// Dependencies: github.com/cjtrowbridge/apparat/internal/hud, github.com/hajimehoshi/ebiten/v2.
package gui

import (
	"image"
	"testing"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestBodyScrollsRegisterForSettingsAndMasterDetail(t *testing.T) {
	game := horizontalGame()
	{
		got := len(game.verticalScrolls)
		if got != 2 {
			t.Fatalf("master-detail vertical scrolls = %d, want 2", got)
		}
	}
	selectTabThroughStrip(t, game, hud.TabSettings)
	{
		got := len(game.verticalScrolls)
		if got != 1 {
			t.Fatalf("settings vertical scrolls = %d, want 1", got)
		}
	}
}

func TestBodyScrollChangesOnlyTheTargetViewport(t *testing.T) {
	game := horizontalGame()
	first := game.verticalScrolls[0]
	selectTabThroughStrip(t, game, hud.TabSettings)
	game.Layout(640, 400)
	game.rebuildUI(game.shell.Snapshot())
	game.ui.Container.SetLocation(image.Rect(0, 0, 640, 400))
	game.ui.Draw(ebiten.NewImage(640, 400))
	second := game.verticalScrolls[0]
	first.ScrollTop = 0.25
	second.ScrollTop = 0.75
	game.bodyScroll = second
	game.scrollBodyBy(-72)
	{
		got := second.ScrollTop
		if got <= 0.75 {
			t.Fatalf("target ScrollTop = %.2f, want greater than 0.75", got)
		}
	}
	{
		got := first.ScrollTop
		if got != 0.25 {
			t.Fatalf("other ScrollTop = %.2f, want unchanged 0.25", got)
		}
	}
}

func TestBodyScrollDragSuppressesReleaseTarget(t *testing.T) {
	game := horizontalGame()
	game.bodyScroll = game.verticalScrolls[0]
	game.scrollBodyBy(-bodyScrollDragThreshold)
	game.finishBodyScroll()
	if !game.bodySelectionSuppressed() {
		t.Fatal("body selection was not suppressed after drag release")
	}
	for range bodyScrollCancelUpdateFrames - 1 {
		game.advanceBodyScrollCancellation()
		if !game.bodySelectionSuppressed() {
			t.Fatal("body selection suppression ended before deferred events")
		}
	}
	game.advanceBodyScrollCancellation()
	if game.bodySelectionSuppressed() {
		t.Fatal("body selection suppression did not end after deferred event window")
	}
}

func TestBodyScrollDragRetainsTabStripPosition(t *testing.T) {
	game := horizontalGame()
	game.tabScroll.ScrollLeft = 0.4
	game.lockBodyTabScroll()
	game.bodyScroll = game.verticalScrolls[0]
	game.scrollBodyBy(-bodyScrollDragThreshold)
	game.finishBodyScroll()
	game.tabScroll.ScrollLeft = 0.8
	game.enforceBodyTabScrollLock()
	{
		got := game.tabScroll.ScrollLeft
		if got != 0.4 {
			t.Fatalf("tab ScrollLeft = %.2f, want body-gesture snapshot 0.40", got)
		}
	}
}

func TestBodyViewportStaysBetweenTabsAndDiagnostics(t *testing.T) {
	for _, width := range []int{360, 1280} {
		game := horizontalGame()
		game.Layout(width, 800)
		game.rebuildUI(game.shell.Snapshot())
		game.ui.Container.SetLocation(image.Rect(0, 0, width, 800))
		game.ui.Draw(ebiten.NewImage(width, 800))
		assertBodyViewports(t, game, width)

		selectTabThroughStrip(t, game, hud.TabSettings)
		game.ui.Container.SetLocation(image.Rect(0, 0, width, 800))
		game.ui.Draw(ebiten.NewImage(width, 800))
		assertBodyViewports(t, game, width)
	}
}

func assertBodyViewports(t *testing.T, game *Game, width int) {
	t.Helper()
	if len(game.verticalScrolls) == 0 {
		t.Fatal("no vertical body scrolls registered")
	}
	minimumY := tabTop + tabHeight + bodyGap
	maximumY := game.height - diagnosticsHeight
	for _, scroll := range game.verticalScrolls {
		view := scroll.ViewRect()
		if view.Min.Y < minimumY {
			t.Fatalf("viewport top = %d, want >= %d at width %d", view.Min.Y, minimumY, width)
		}
		if view.Max.Y > maximumY {
			t.Fatalf("viewport bottom = %d, want <= %d at width %d", view.Max.Y, maximumY, width)
		}
	}
}
