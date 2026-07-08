//go:build gui

package gui

import "testing"

func TestTabIndexAtUsesDrawnTabRects(t *testing.T) {
	game := &Game{tabRects: []tabRect{{index: 2, x: windowMargin, y: tabTop, w: tabWideWidth, h: tabHeight}}}
	index, ok := game.tabIndexAt(windowMargin+20, tabTop+20)
	if !ok || index != 2 {
		t.Fatalf("hit = index %d ok %t, want index 2 ok true", index, ok)
	}
	if _, ok := game.tabIndexAt(4, 4); ok {
		t.Fatal("unexpected hit outside tab rect")
	}
}
