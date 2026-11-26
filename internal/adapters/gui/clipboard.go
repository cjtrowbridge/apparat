//go:build gui

// Routes clipboard shortcuts to the focused editor and reports transfer failures.
// Clipboard payloads stay in the operating system and the active editing transaction.
package gui

import (
	"image/color"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type clipboardNotice struct {
	message string
	until   time.Time
}

func (g *Game) clipboardError(message string) {
	g.clipboardNotice = clipboardNotice{message, time.Now().Add(4 * time.Second)}
}

func (g *Game) drawClipboardNotice(dst *ebiten.Image) {
	n := g.clipboardNotice
	if time.Now().Before(n.until) {
		vector.FillRect(dst, 0, 0, float32(dst.Bounds().Dx()), 28, color.Black, false)
		ebitenutil.DebugPrintAt(dst, n.message, 8, 5)
	}
}

func (g *Game) copyDocument(value string) bool {
	if value == "" {
		return false
	}
	copyText := g.runtimeInfo.CopyText
	if copyText == nil {
		copyText = systemCopy
	}
	if !copyText(value) {
		g.clipboardError("Copy failed. Check system clipboard access.")
		return false
	}
	return true
}

func (g *Game) pasteInto(insert func(string)) {
	paste := g.runtimeInfo.PasteText
	if paste == nil {
		paste = systemPaste
	}
	value := paste()
	if value == "" || !utf8.ValidString(value) || strings.ContainsRune(value, 0) {
		g.clipboardError("Paste needs text from the system clipboard.")
		return
	}
	insert(value)
}

func shortcutModifier() bool {
	return ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
}

// The pinned EbitenUI exposes selected bytes but not the rune offsets. Its
// SelectedText and DeleteSelectedText slice UTF-8 with rune indexes. Read only
// the two indexes here and apply changes through its public validated setter.
func fieldSelection(field *widget.TextInput) (int, int, bool) {
	v := reflect.ValueOf(field).Elem()
	anchor, caret := v.FieldByName("dragStartIndex"), v.FieldByName("cursorPosition")
	if anchor.Kind() != reflect.Int || caret.Kind() != reflect.Int {
		return 0, 0, false
	}
	a, b := int(anchor.Int()), int(caret.Int())
	if a < 0 {
		a = b
	}
	count := utf8.RuneCountInString(field.GetText())
	return min(a, b), max(a, b), a >= 0 && b >= 0 && a <= count && b <= count
}

func (g *Game) fieldClipboardKeys() {
	if !shortcutModifier() {
		return
	}
	field, ok := g.ui.GetFocusedWidget().(*widget.TextInput)
	if !ok {
		p := g.bubbleText.active
		if p != nil && g.actions.mode == "" && inpututil.IsKeyJustPressed(ebiten.KeyC) {
			g.copyDocument(p.selected())
		}
		return
	}
	if field.GetWidget().Disabled {
		return
	}
	_, ordinary := field.GetWidget().CustomData.(editorIdentity)
	if !ordinary {
		return
	}
	for _, key := range []ebiten.Key{ebiten.KeyA, ebiten.KeyC, ebiten.KeyX, ebiten.KeyV} {
		if inpututil.IsKeyJustPressed(key) {
			g.fieldClipboard(field, key)
			return
		}
	}
}

func (g *Game) fieldClipboard(field *widget.TextInput, key ebiten.Key) {
	if key == ebiten.KeyA {
		field.SelectAll()
		return
	}
	a, b, ok := fieldSelection(field)
	if !ok {
		g.clipboardError("Text selection unavailable.")
		return
	}
	runes := []rune(field.GetText())
	replace := func(value string) {
		// A physical paste into a single-line input never submits the form.
		value = searchLine(value)
		candidate := string(runes[:a]) + value + string(runes[b:])
		field.SetText(candidate)
		field.CursorMoveEnd()
		for range len(runes) - b {
			field.CursorMoveLeft()
		}
	}
	switch key {
	case ebiten.KeyC:
		g.copyDocument(string(runes[a:b]))
	case ebiten.KeyX:
		if g.copyDocument(string(runes[a:b])) {
			replace("")
		}
	case ebiten.KeyV:
		g.pasteInto(replace)
	}
}
