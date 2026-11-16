//go:build gui

// Verifies checks bounded tab-strip scrolling.
// Uses package-local contracts and Go standard-library types.
package gui

import "testing"

func TestTabStripRebuildLeftAlignsWhenTabsFit(t *testing.T) {
	game := NewGame()
	game.width = game.tabStripContentWidth + (windowMargin * 2)
	game.rebuildUI(game.shell.Snapshot())
	game.tabScroll.ScrollLeft = 0.65
	game.rebuildUI(game.shell.Snapshot())
	{
		got := game.tabScroll.ScrollLeft
		if got != 0 {
			t.Fatalf("tab strip scroll after fitting rebuild = %.2f, want 0", got)
		}
	}
}

func TestTabStripRebuildPreservesOverflowDragPosition(t *testing.T) {
	game := NewGame()
	game.width = 280
	game.rebuildUI(game.shell.Snapshot())
	if !game.tabStripOverflows() {
		t.Fatal("narrow tab strip should overflow")
	}
	game.tabScroll.ScrollLeft = 0.42
	game.rebuildUI(game.shell.Snapshot())
	{
		got := game.tabScroll.ScrollLeft
		if got != 0.42 {
			t.Fatalf("overflow tab strip scroll after rebuild = %.2f, want preserved 0.42", got)
		}
	}
}

func TestPointerTabSelectionPreservesOverflowDragPosition(t *testing.T) {
	game := NewGame()
	game.width = 280
	game.rebuildUI(game.shell.Snapshot())
	game.tabScroll.ScrollLeft = 0.42
	game.selectTab(1, game.tabButtons[1])
	{
		got := game.tabScroll.ScrollLeft
		if got != 0.42 {
			t.Fatalf("pointer-selected tab strip scroll = %.2f, want preserved 0.42", got)
		}
	}
}

func TestRequestedTabRevealMovesOnlyWhenSelectedTabIsClipped(t *testing.T) {
	game := NewGame()
	game.width = 280
	game.rebuildUI(game.shell.Snapshot())
	game.selectTabWithReveal(len(game.shell.Snapshot().Tabs) - 1)
	game.rebuildUI(game.shell.Snapshot())
	game.ensureActiveTabVisible()
	{
		got := game.tabScroll.ScrollLeft
		if got != 1 {
			t.Fatalf("last clipped tab scroll = %.2f, want minimal right edge 1", got)
		}
	}
}

func TestTabHoverImageDiffersFromSelectedImage(t *testing.T) {
	image := tabButtonImage()
	if image.Hover == image.Pressed {
		t.Fatal("tab hover image matches selected image")
	}
}

func TestIconTabsFitPhoneWidthWithTouchTargetsAndNames(t *testing.T) {
	game := NewGame()
	game.width = 320
	game.rebuildUI(game.shell.Snapshot())
	if game.tabStripOverflows() {
		t.Fatal("all five icon tabs must fit at 320 logical pixels")
	}
	for index, button := range game.tabButtons {
		width, height := button.PreferredSize()
		if width < 44 || height < 44 || width != tabIconButtonWidth {
			t.Fatalf("tab %d size %dx%d violates compact touch target", index, width, height)
		}
		if len(button.GetWidget().ToolTips) == 0 {
			t.Fatalf("tab %d lacks a name tooltip", index)
		}
		if index > 0 && button.GraphicImage().Idle == nil {
			t.Fatalf("tab %d lacks embedded artwork", index)
		}
	}
}
