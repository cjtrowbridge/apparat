//go:build gui

// Verifies checks bounded tab-strip scrolling.
// Uses package-local contracts and Go standard-library types.
package gui

import "testing"

func TestTabStripRebuildLeftAlignsWhenTabsFit(t *testing.T) {
	game := horizontalGame()
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
	game := horizontalGame()
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
	game := horizontalGame()
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
	game := horizontalGame()
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

func TestScrollablePhoneTabsKeepTouchTargetsAndNames(t *testing.T) {
	game := horizontalGame()
	game.width = 320
	game.rebuildUI(game.shell.Snapshot())
	if !game.tabStripOverflows() {
		t.Fatal("seven touch targets require horizontal scrolling at 320 logical pixels")
	}
	for index, button := range game.tabButtons {
		width, height := button.PreferredSize()
		if width < 44 || height < 44 {
			t.Fatalf("tab %d size %dx%d violates compact touch target", index, width, height)
		}
		if len(button.GetWidget().ToolTips) != 0 {
			t.Fatalf("tab %d has an unwanted tooltip", index)
		}
		if index > 0 && button.GraphicImage().Idle == nil {
			t.Fatalf("tab %d lacks embedded artwork", index)
		}
	}
}

func TestSixTabsFitMinimumPhoneWidth(t *testing.T) {
	const available = 320 - 2*windowMargin
	width := tabTargetWidth(available, 6)
	if width < 44 || width*6+5*tabGap > available {
		t.Fatalf("six targets of width %d do not fit %d", width, available)
	}
}

func TestResponsivePlacementAdaptsWithoutChangingSavedChoice(t *testing.T) {
	game := horizontalGame()
	p := game.prefs.Preferences
	p.Placement = "responsive"
	game.applyPreferences(p)
	for _, tc := range []struct {
		width, height int
		rail          string
	}{
		{1100, 800, "left_rail"}, {390, 804, ""}, {1100, 340, ""}, {1000, 620, "left_rail"},
	} {
		game.Layout(tc.width, tc.height)
		game.rebuildUI(game.shell.Snapshot())
		if game.sideRail() != tc.rail || (game.tabScroll == nil) != (tc.rail != "") || game.prefs.Placement != "responsive" {
			t.Fatalf("incorrect navigation for %dx%d", tc.width, tc.height)
		}
	}
}

func TestRightRailClearsPushToTalk(t *testing.T) {
	game := horizontalGame()
	game.Layout(640, 600)
	game.prefs.Placement = "right_rail"
	x, _ := game.pttCenter()
	if x+pttRadius >= game.width-windowMargin-tabIconButtonWidth {
		t.Fatal("push-to-talk overlaps the right rail")
	}
}

func TestBottomTabsKeepPushToTalkAboveStrip(t *testing.T) {
	game := horizontalGame()
	game.prefs.Placement = "bottom"
	for _, width := range []int{320, 1024} {
		game.Layout(width, 600)
		game.rebuildUI(game.shell.Snapshot())
		_, cy := game.pttCenter()
		if game.tabScroll == nil || cy+pttRadius >= game.height-diagnosticsHeight-tabHeight {
			t.Fatalf("bottom strip obscured at width %d", width)
		}
	}
}
