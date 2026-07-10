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

func TestLocalRectConversionUsesPaneOrigin(t *testing.T) {
	pane := rect{x: 100, y: 200, w: 300, h: 400}
	child := rect{x: 112, y: 244, w: 80, h: 44}
	local := child.localTo(pane)
	if local.x != 12 || local.y != 44 || local.w != child.w || local.h != child.h {
		t.Fatalf("local rect = %+v, want x=12 y=44 w=%d h=%d", local, child.w, child.h)
	}
}

func TestMasterDetailTextRectsAreInsideExpectedPanes(t *testing.T) {
	body := tabBodyRect(1280, 800)
	list, detail := masterDetailRects(body)
	listText := rect{x: list.x + fieldsetPadding, y: list.y + fieldsetPadding + 28, w: list.w - fieldsetPadding*2, h: touchTargetH}
	detailText := rect{x: detail.x + fieldsetPadding, y: detail.y + fieldsetPadding, w: detail.w - fieldsetPadding*2, h: fieldsetDescH}
	if !list.contains(listText.x, listText.y) || !list.contains(listText.x+listText.w-1, listText.y+listText.h-1) {
		t.Fatalf("list text rect %+v outside list pane %+v", listText, list)
	}
	if !detail.contains(detailText.x, detailText.y) || !detail.contains(detailText.x+detailText.w-1, detailText.y+detailText.h-1) {
		t.Fatalf("detail text rect %+v outside detail pane %+v", detailText, detail)
	}
}

func TestSettingsUpdatesSectionIsFirst(t *testing.T) {
	tab := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[6]
	first := tab.Sections[0]
	if first.Title != "Updates" {
		t.Fatalf("first settings section = %q, want Updates", first.Title)
	}
	if len(first.Rows) != 0 {
		t.Fatalf("updates rows = %d, want native button to own control row", len(first.Rows))
	}
}

func TestTruncateTextUsesASCIIEllipsis(t *testing.T) {
	got := truncateText("abcdefghijklmnopqrstuvwxyz", debugGlyphWidth*8)
	if got != "abcde..." {
		t.Fatalf("truncateText = %q, want ASCII ellipsis", got)
	}
}

func TestWrapTextKeepsTextInsideBlockWidth(t *testing.T) {
	lines := wrapText("alpha beta gamma delta", debugGlyphWidth*11)
	if len(lines) < 2 {
		t.Fatalf("wrapText lines = %+v, want multiple lines", lines)
	}
	for _, line := range lines {
		if labelWidth(line) > debugGlyphWidth*11 {
			t.Fatalf("line %q exceeds block width", line)
		}
	}
}

func TestClampScroll(t *testing.T) {
	if got := clampScroll(-5, 200); got != 0 {
		t.Fatalf("negative scroll = %d, want 0", got)
	}
	if got := clampScroll(300, 200); got != 200 {
		t.Fatalf("overscroll = %d, want 200", got)
	}
	if got := clampScroll(20, -1); got != 0 {
		t.Fatalf("short content scroll = %d, want 0", got)
	}
}

func TestDragBodyPaneStartsSmoothlyAfterThreshold(t *testing.T) {
	game := NewGame()
	game.width = 420
	game.height = 320
	state := bodyDragState{active: true, pane: scrollPaneSettings, startY: 100, startScroll: 50}
	game.bodyScroll.settings = 50
	game.dragBodyPane(&state, 100+bodyDragThreshold+1)
	if !state.dragged {
		t.Fatal("body drag should start after threshold")
	}
	if game.bodyScroll.settings != 50 {
		t.Fatalf("settings scroll = %d, want no threshold jump", game.bodyScroll.settings)
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

func TestNativeUpdateSlotUsesStableID(t *testing.T) {
	button, ok := nativeControlSlotRect(nativeSlotUpdate, 1600, 1260)
	if !ok {
		t.Fatal("update native slot missing")
	}
	if _, ok := nativeControlSlotRect("missing.slot", 1600, 1260); ok {
		t.Fatal("unexpected native slot for unknown id")
	}
	if button.h < touchTargetH || button.w <= button.h {
		t.Fatalf("button slot should be touch-sized and wider than tall: %+v", button)
	}
}

func TestInputPlaceholderRectStaysInsideFieldset(t *testing.T) {
	fieldset := rect{x: 10, y: 20, w: 240, h: 160}
	input := inputPlaceholderRect(fieldset, 1)
	if input.x < fieldset.x || input.x+input.w > fieldset.x+fieldset.w {
		t.Fatalf("input rect %+v outside fieldset %+v", input, fieldset)
	}
	if input.h < touchTargetH {
		t.Fatalf("input height = %d, want at least %d", input.h, touchTargetH)
	}
}

func TestLiveUpdateButtonSlotMovesWithSettingsScroll(t *testing.T) {
	game := NewGame()
	if err := game.shell.SelectTab(6); err != nil {
		t.Fatal(err)
	}
	game.bodyScroll.settings = 30
	live := game.UpdateButtonY(1280, 800)
	static := UpdateButtonY(1280, 800)
	if live != static-30 {
		t.Fatalf("live button y = %d, want static y %d minus scroll", live, static)
	}
}

func TestLiveUpdateButtonSlotHidesWhenScrolledOutOfView(t *testing.T) {
	game := NewGame()
	if err := game.shell.SelectTab(6); err != nil {
		t.Fatal(err)
	}
	if !game.UpdateButtonVisible(1280, 800) {
		t.Fatal("update button should be visible before Settings scroll")
	}
	game.bodyScroll.settings = 10000
	if game.UpdateButtonVisible(1280, 800) {
		t.Fatal("update button should be hidden after its slot scrolls out of the body")
	}
}

func TestBodyRowsUseTouchTargetHeight(t *testing.T) {
	if fieldsetRowH < 44 || touchTargetH < 44 {
		t.Fatalf("touch target sizes too small: row=%d target=%d", fieldsetRowH, touchTargetH)
	}
}
