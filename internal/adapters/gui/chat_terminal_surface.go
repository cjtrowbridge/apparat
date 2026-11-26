//go:build gui

// Draws the selected PTY screen and sends focused keyboard input to its controller.
package gui

import (
	"fmt"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image"
	"image/color"
)

type chatTerminalSurface struct {
	*widget.TextInput
	g       *Game
	id      string
	address modelAddress
	t       *chatTerminalPane
	output  *terminalOutput
	cell    float64
	line    int
	ime     textinput.Composer
}

func (g *Game) newChatTerminalSurface(a modelAddress, t *chatTerminalPane) *chatTerminalSurface {
	e := &chatTerminalSurface{TextInput: g.newProjectInput("", ignoreDocumentTextChange), g: g, address: a, id: t.selected, t: t}
	e.GetWidget().CustomData = e
	size := float64(max(10, min(16, g.prefs.Font)))
	e.cell = text.Advance("M", mdStyle{mono: true, size: size}.face())
	e.line = terminalLineHeight(size)
	e.ime.OnNewSession = func() *textinput.SessionOptions {
		r := e.GetWidget().Rect.Inset(4)
		return &textinput.SessionOptions{CaretBounds: image.Rect(r.Min.X, r.Min.Y, r.Min.X+2, r.Min.Y+e.line)}
	}
	e.ime.OnCommit = func(c *textinput.Commit) {
		e.enqueue(c.Text())
	}
	e.ime.OnEndByUser = func() {
		g.ui.ClearFocus()
	}
	return e
}
func (e *chatTerminalSurface) PreferredSize() (int, int) {
	return 44, 44
}
func (e *chatTerminalSurface) enqueue(s string) {
	t := e.t
	if s == "" || e.id != t.selected {
		return
	}
	if t.queueID != t.selected {
		t.queue = ""
		t.queueID = t.selected
	}
	if len(t.queue)+len(s) > 65536 {
		t.notice = "Terminal input buffer full."
		e.g.layoutDirty = true
		return
	}
	if e.output != nil && e.g.bubbleText.active == &e.output.passage {
		e.g.bubbleText.clear()
	}
	t.queue += s
}
func (e *chatTerminalSurface) Render(screen *ebiten.Image) {
	r := e.GetWidget().Rect
	clip := screen.Bounds().Intersect(r)
	if clip.Empty() {
		return
	}
	dst := screen.SubImage(clip).(*ebiten.Image)
	dst.Fill(panelColor)
	if e.output == nil {
		return
	}
	e.output.update(e.t.screen, r.Dx())
	e.output.passage.layout.paint(dst, r.Inset(4))
	if e.output.frozen() {
		return
	}
	var col, row int
	_, err := fmt.Sscanf(e.t.caret, "%d,%d", &col, &row)
	if err == nil {
		vector.FillRect(dst, float32(r.Min.X+4)+float32(float64(col)*e.cell), float32(r.Min.Y+4+(row+1)*e.line-2), float32(e.cell), 2, color.White, false)
	}

}
func (e *chatTerminalSurface) Update(info *widget.UpdateObject) {
	e.GetWidget().Update(info)
	r := e.GetWidget().Rect
	if r.Dx() > 0 && r.Dy() > 0 && e.t.pending == nil && e.t.notice == "" {
		size := fmt.Sprintf("%dx%d", max(20, min(200, int(float64(r.Dx()-8)/e.cell))), max(5, min(80, (r.Dy()-8)/e.line)))
		if size != e.t.size {
			e.g.requestChatTerminal(e.address, e.t, "resize", e.t.selected, size)
			e.t.size = size
		}
	}
	if !e.IsFocused() || e.t.permissionPage {
		e.ime.Cancel()
		return
	}
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl)
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyBracketRight) {
		e.ime.Cancel()
		e.g.ui.ClearFocus()
		e.g.setNavTarget(e)
		return
	}
	if ctrl {
		e.controlKeys(ebiten.IsKeyPressed(ebiten.KeyShift), inpututil.IsKeyJustPressed)
		return
	}
	for _, key := range []struct {
		k     ebiten.Key
		value string
	}{
		{ebiten.KeyBackspace, "\x7f"}, {ebiten.KeyEnter, "\r"}, {ebiten.KeyTab, "\t"}, {ebiten.KeyEscape, "\x1b"},
		{ebiten.KeyArrowUp, "\x1b[A"}, {ebiten.KeyArrowDown, "\x1b[B"}, {ebiten.KeyArrowRight, "\x1b[C"}, {ebiten.KeyArrowLeft, "\x1b[D"},
		{ebiten.KeyHome, "\x1b[H"}, {ebiten.KeyEnd, "\x1b[F"}, {ebiten.KeyDelete, "\x1b[3~"},
	} {
		if navigationRepeat(inpututil.KeyPressDuration(key.k)) {
			e.enqueue(key.value)
			return
		}
	}
	handled, _ := e.ime.Update()
	if !handled {
		e.enqueue(string(ebiten.AppendInputChars(nil)))
	}
}

func (e *chatTerminalSurface) controlKeys(shift bool, pressed func(ebiten.Key) bool) {
	if shift {
		if pressed(ebiten.KeyC) {
			if e.output != nil {
				e.output.copySelection()
			}
			return
		}
		if pressed(ebiten.KeyV) {
			e.paste()
			return
		}
	}
	for k := ebiten.KeyA; k <= ebiten.KeyZ; k++ {
		if pressed(k) {
			e.enqueue(string(byte(k - ebiten.KeyA + 1)))
		}
	}
}

func (g *Game) terminalHasFocus() bool {
	f, ok := g.ui.GetFocusedWidget().(*widget.TextInput)
	if !ok {
		return false
	}
	_, ok = f.GetWidget().CustomData.(*chatTerminalSurface)
	return ok
}
