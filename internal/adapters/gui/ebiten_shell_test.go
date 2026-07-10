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
