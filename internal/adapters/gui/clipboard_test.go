//go:build gui

// Exercises focused field replacement, transfer failures and rendered passage offsets.
// All clipboard content in these tests is synthetic.
package gui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"strings"
	"testing"
)

func TestClipboardFieldRanges(t *testing.T) {
	g := NewGame()
	defer g.Close()
	f := g.newProjectInput("é foo foo", ignoreDocumentTextChange)
	f.CursorMoveEnd()
	for range 3 {
		f.CursorMoveLeft()
	}
	g.runtimeInfo.PasteText = func() string {
		return "🙂"
	}
	g.fieldClipboard(f, ebiten.KeyV)
	if f.GetText() != "é foo 🙂foo" {
		t.Fatal("paste at Unicode caret", f.GetText())
	}
	f.SelectAll()
	g.runtimeInfo.CopyText = func(value string) bool {
		if value != f.GetText() {
			t.Fatal("inexact Unicode copy")
		}
		return false
	}
	g.fieldClipboard(f, ebiten.KeyX)
	a, b, ok := fieldSelection(f)
	if !ok || a != 0 || b != 10 || f.GetText() != "é foo 🙂foo" {
		t.Fatal("failed cut mutated field", a, b)
	}
	g.runtimeInfo.PasteText = func() string {
		return ""
	}
	g.fieldClipboard(f, ebiten.KeyV)
	if f.GetText() != "é foo 🙂foo" {
		t.Fatal("empty paste erased selection")
	}
}

func TestClipboardMarkdownOffsets(t *testing.T) {
	for _, width := range []int{46, 480} {
		m := layoutMarkdown("**café** foo foo 🙂  \nnext\n\n```\n  x\n    y\n```", width, 14)
		got := strings.TrimSuffix(m.plain.String(), "\n")
		if got != "café foo foo 🙂\nnext\n  x\n    y" {
			t.Fatalf("%d: %q", width, got)
		}
		for _, r := range m.runs {
			if r.fill == nil && m.plain.String()[r.start:r.start+len(r.value)] != r.value {
				t.Fatal("run offset", r)
			}
		}
	}
}
