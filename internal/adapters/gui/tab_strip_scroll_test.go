//go:build gui

package gui

import "testing"

func TestTabStripRebuildLeftAlignsWhenTabsFit(t *testing.T) {
	game := NewGame()
	game.width = game.tabStripContentWidth + (windowMargin * 2)
	game.rebuildUI(game.shell.Snapshot())
	game.tabScroll.ScrollLeft = 0.65
	game.rebuildUI(game.shell.Snapshot())
	if got := game.tabScroll.ScrollLeft; got != 0 {
		t.Fatalf("tab strip scroll after fitting rebuild = %.2f, want 0", got)
	}
}

func TestTabStripRebuildPreservesOverflowDragPosition(t *testing.T) {
	game := NewGame()
	game.width = 360
	game.rebuildUI(game.shell.Snapshot())
	if !game.tabStripOverflows() {
		t.Fatal("narrow tab strip should overflow")
	}
	game.tabScroll.ScrollLeft = 0.42
	game.rebuildUI(game.shell.Snapshot())
	if got := game.tabScroll.ScrollLeft; got != 0.42 {
		t.Fatalf("overflow tab strip scroll after rebuild = %.2f, want preserved 0.42", got)
	}
}

func TestPointerTabSelectionPreservesOverflowDragPosition(t *testing.T) {
	game := NewGame()
	game.width = 360
	game.rebuildUI(game.shell.Snapshot())
	game.tabScroll.ScrollLeft = 0.42
	game.selectTab(1, game.tabButtons[1])
	if got := game.tabScroll.ScrollLeft; got != 0.42 {
		t.Fatalf("pointer-selected tab strip scroll = %.2f, want preserved 0.42", got)
	}
}

func TestRequestedTabRevealMovesOnlyWhenSelectedTabIsClipped(t *testing.T) {
	game := NewGame()
	game.width = 360
	game.rebuildUI(game.shell.Snapshot())
	game.selectTabWithReveal(5)
	game.rebuildUI(game.shell.Snapshot())
	game.ensureActiveTabVisible()
	if got := game.tabScroll.ScrollLeft; got != 1 {
		t.Fatalf("last clipped tab scroll = %.2f, want minimal right edge 1", got)
	}
}

func TestTabHoverImageDiffersFromSelectedImage(t *testing.T) {
	image := tabButtonImage()
	if image.Hover == image.Pressed {
		t.Fatal("tab hover image matches selected image")
	}
}
