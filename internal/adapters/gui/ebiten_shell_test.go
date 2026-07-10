//go:build gui

package gui

import (
	"testing"

	"github.com/cjtrowbridge/apparat/internal/hud"
)

func TestTabIndexAtUsesDrawnTabRects(t *testing.T) {
	game := &Game{width: 320, tabRects: []tabRect{{index: 2, x: windowMargin, y: tabTop, w: 90, h: tabHeight}}}
	index, ok := game.tabIndexAt(windowMargin+20, tabTop+20)
	if !ok || index != 2 {
		t.Fatalf("hit = index %d ok %t, want index 2 ok true", index, ok)
	}
	if _, ok := game.tabIndexAt(4, 4); ok {
		t.Fatal("unexpected hit outside tab rect")
	}
}

func TestTabButtonWidthUsesLargestLabelWithBalancedPadding(t *testing.T) {
	snapshot := hud.Snapshot{Tabs: []hud.Tab{
		{Descriptor: hud.TabDescriptor{Label: "A", Glyph: "*"}},
		{Descriptor: hud.TabDescriptor{Label: "Longest", Glyph: "*"}},
	}}
	got := tabButtonWidth(snapshot)
	want := labelWidth("* Longest") + tabTextInsetX*2
	if got != want {
		t.Fatalf("tabButtonWidth = %d, want %d", got, want)
	}
}

func TestClampTabScroll(t *testing.T) {
	if got := clampTabScroll(-20, 300, 100); got != 0 {
		t.Fatalf("negative scroll = %d, want 0", got)
	}
	if got := clampTabScroll(500, 300, 100); got != 200 {
		t.Fatalf("overscroll = %d, want 200", got)
	}
	if got := clampTabScroll(40, 80, 100); got != 0 {
		t.Fatalf("short content scroll = %d, want 0", got)
	}
}

func TestEnsureTabVisibleScrollsActiveTabIntoViewport(t *testing.T) {
	game := NewGame()
	game.width = 180
	if err := game.shell.SelectTab(6); err != nil {
		t.Fatal(err)
	}
	game.ensureTabVisible(6)
	if game.tabScrollX <= 0 {
		t.Fatalf("tabScrollX = %d, want positive scroll", game.tabScrollX)
	}
	if game.tabScrollX != clampTabScroll(game.tabScrollX, game.tabContentW, game.tabViewportWidth()) {
		t.Fatalf("tabScrollX = %d outside clamp", game.tabScrollX)
	}
}

func TestDragTabsDoesNotScrollBeforeThreshold(t *testing.T) {
	game := NewGame()
	game.width = 180
	game.tabContentW = 600
	state := dragState{active: true, startX: 100, startScroll: 50}
	game.dragTabs(&state, 100+tabDragThreshold)
	if state.dragged {
		t.Fatal("drag should not start at the threshold")
	}
	if game.tabScrollX != 0 {
		t.Fatalf("tabScrollX = %d, want unchanged zero before drag starts", game.tabScrollX)
	}
}

func TestDragTabsStartsSmoothlyAfterThreshold(t *testing.T) {
	game := NewGame()
	game.width = 180
	game.tabContentW = 600
	game.tabScrollX = 50
	state := dragState{active: true, startX: 100, startScroll: 50}
	game.dragTabs(&state, 100+tabDragThreshold+1)
	if !state.dragged {
		t.Fatal("drag should start after threshold")
	}
	if game.tabScrollX != 50 {
		t.Fatalf("tabScrollX = %d, want no threshold jump", game.tabScrollX)
	}
	game.dragTabs(&state, 130)
	if game.tabScrollX >= 50 {
		t.Fatalf("tabScrollX = %d, want smooth incremental movement below 50", game.tabScrollX)
	}
}

func TestTabPressCanSelectBeforeDragStarts(t *testing.T) {
	game := NewGame()
	game.width = 320
	game.tabRects = []tabRect{{index: 1, x: windowMargin, y: tabTop, w: 90, h: tabHeight}}
	index, ok := game.tabIndexAt(windowMargin+20, tabTop+20)
	if !ok || index != 1 {
		t.Fatalf("press target = index %d ok %t, want index 1 ok true", index, ok)
	}
}

func TestMasterDetailRectsStaySeparatedAtNarrowWidth(t *testing.T) {
	list, detail := masterDetailRects(rect{x: 0, y: 0, w: 420, h: 300})
	if list.w <= 0 || detail.w <= 0 {
		t.Fatalf("invalid pane widths: list=%+v detail=%+v", list, detail)
	}
	if list.x+list.w+masterDividerW > detail.x {
		t.Fatalf("panes overlap: list=%+v detail=%+v", list, detail)
	}
}

func TestSettingsUpdatesSectionIsLast(t *testing.T) {
	tab := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[6]
	last := tab.Sections[len(tab.Sections)-1]
	if last.Title != "Updates" {
		t.Fatalf("last settings section = %q, want Updates", last.Title)
	}
	if len(last.Rows) != 0 {
		t.Fatalf("updates rows = %d, want native button to own control row", len(last.Rows))
	}
}

func TestTruncateTextUsesASCIIEllipsis(t *testing.T) {
	got := truncateText("abcdefghijklmnopqrstuvwxyz", debugGlyphWidth*8)
	if got != "abcde..." {
		t.Fatalf("truncateText = %q, want ASCII ellipsis", got)
	}
}

func TestUpdateButtonRectSitsInsideSettingsBody(t *testing.T) {
	body := tabBodyRect(1600, 1260)
	button := updateButtonRect(1600, 1260)
	if button.x < body.x || button.y < body.y || button.x+button.w > body.x+body.w || button.y+button.h > body.y+body.h {
		t.Fatalf("button rect %+v outside body %+v", button, body)
	}
	if button.h < touchTargetH {
		t.Fatalf("button height = %d, want at least %d", button.h, touchTargetH)
	}
}

func TestBodyRowsUseTouchTargetHeight(t *testing.T) {
	if fieldsetRowH < 44 || touchTargetH < 44 {
		t.Fatalf("touch target sizes too small: row=%d target=%d", fieldsetRowH, touchTargetH)
	}
}
