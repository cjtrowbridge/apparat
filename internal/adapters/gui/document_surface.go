//go:build gui

// Renders and edits UTF-8 documents on the game canvas. Positions are byte offsets.
// The embedded focus widget integrates with HUD navigation and the shared PTT target.
package gui

import (
	"image"
	"image/color"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/cjtrowbridge/apparat/internal/domain"
	uiinput "github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type documentSurface struct {
	*widget.TextInput
	game                             *Game
	panel                            *workspacePanel
	body                             string
	starts                           []int
	anchor, caret, top, left, height int
	drag                             bool
	ime                              textinput.Composer
	undo, redo                       []string
	composing                        string
	imeBegin, imeEnd                 int
	committing                       bool
}

func (g *Game) newDocumentSurface(w *workspacePanel) *documentSurface {
	e := &documentSurface{TextInput: g.newProjectInput("", func(string) {}), game: g, panel: w, height: max(20, g.prefs.Font+5)}
	e.GetWidget().MinHeight = 240
	e.load(strings.Join(w.lines, "\n"))
	e.setupIME()
	return e
}
func (e *documentSurface) load(s string) {
	e.body = s
	e.TextInput.SetText(s)
	e.starts = []int{0}
	for i, r := range s {
		if r == '\n' {
			e.starts = append(e.starts, i+1)
		}
	}
	e.caret = min(e.caret, len(s))
	for e.caret > 0 && e.caret < len(s) && !utf8.RuneStart(s[e.caret]) {
		e.caret--
	}
	e.anchor = min(e.anchor, len(s))
	for e.anchor > 0 && e.anchor < len(s) && !utf8.RuneStart(s[e.anchor]) {
		e.anchor--
	}
	e.syncIME()
}
func (e *documentSurface) syncIME() {
	if !e.committing {
		e.ime.Cancel()
	}
}
func (e *documentSurface) row(pos int) int {
	return sort.SearchInts(e.starts, pos+1) - 1
}
func (e *documentSurface) line(row int) string {
	end := len(e.body)
	if row+1 < len(e.starts) {
		end = e.starts[row+1] - 1
	}
	return e.body[e.starts[row]:end]
}
func (e *documentSurface) width(s string) int {
	w, _ := text.Measure(s, *e.game.theme.ButtonTheme.TextFace, 0)
	return int(w + 0.5)
}
func (e *documentSurface) pointPos(x, y int) int {
	r := e.GetWidget().Rect.Inset(4)
	row := max(0, min(len(e.starts)-1, e.top+(y-r.Min.Y)/e.height))
	line := e.line(row)
	target := x - r.Min.X + e.left
	runes := []rune(line)
	i := sort.Search(len(runes), func(i int) bool {
		return e.width(string(runes[:i+1])) > target
	})
	return e.starts[row] + len(string(runes[:i]))
}
func (e *documentSurface) reveal() {
	r := e.GetWidget().Rect.Inset(4)
	row := e.row(e.caret)
	e.top = max(0, min(row, max(e.top, row-max(1, r.Dy()/e.height)+1)))
	x := e.width(e.body[e.starts[row]:e.caret])
	e.left = max(0, min(x, max(e.left, x-r.Dx()+3)))
}
func (e *documentSurface) change(s string, pos int) {
	e.caret = pos
	e.anchor = pos
	e.load(s)
	e.panel.lines = strings.Split(s, "\n")
	e.panel.confirmation = ""
	e.game.keepEditingDraft(e.panel)
	e.reveal()
}
func (e *documentSurface) insert(s string) {
	if !e.panel.editing || e.panel.readOnly || e.game.projects.pending != nil {
		return
	}
	a, b := min(e.caret, e.anchor), max(e.caret, e.anchor)
	next := e.body[:a] + s + e.body[b:]
	if len(next) > domain.WorkspaceTextLimit || strings.ContainsRune(s, 0) || !utf8.ValidString(s) {
		return
	}
	if next == e.body {
		return
	}
	e.undo = append(e.undo, e.body)
	if len(e.undo) > 40 {
		e.undo = e.undo[1:]
	}
	e.redo = nil
	e.change(next, a+len(s))
}
func (e *documentSurface) history(from, to *[]string) {
	if len(*from) == 0 {
		return
	}
	*to = append(*to, e.body)
	last := len(*from) - 1
	s := (*from)[last]
	*from = (*from)[:last]
	e.change(s, min(e.caret, len(s)))
}
func (e *documentSurface) Render(screen *ebiten.Image) {
	rect := e.GetWidget().Rect
	clip := rect.Intersect(screen.Bounds())
	if clip.Empty() {
		return
	}
	dst := screen.SubImage(clip).(*ebiten.Image)
	dst.Fill(panelColor)
	a, b := min(e.anchor, e.caret), max(e.anchor, e.caret)
	for i := e.top; i < len(e.starts); i++ {
		y := rect.Min.Y + 4 + (i-e.top)*e.height
		if y >= rect.Max.Y {
			break
		}
		line := e.line(i)
		begin, end := e.starts[i], e.starts[i]+len(line)
		x := rect.Min.X + 4 - e.left
		if b > begin && a <= end && b > a {
			l := e.width(e.body[begin:max(begin, a)])
			h := e.width(e.body[begin:min(b, end)])
			vector.FillRect(dst, float32(x+l), float32(y), float32(max(2, h-l)), float32(e.height), activeTabColor, false)
		}
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		op.ColorScale.ScaleWithColor(color.White)
		if e.composing != "" && e.caret >= begin && e.caret <= end {
			line = e.body[begin:e.caret] + e.composing + e.body[e.caret:end]
		}
		text.Draw(dst, line, *e.game.theme.ButtonTheme.TextFace, op)
		if e.IsFocused() && e.panel.editing && e.caret >= begin && e.caret <= end {
			cx := x + e.width(e.body[begin:e.caret])
			vector.FillRect(dst, float32(cx), float32(y), 2, float32(e.height), color.White, false)
		}
	}
}
func (e *documentSurface) Update(info *widget.UpdateObject) {
	e.GetWidget().Update(info)
	r := e.GetWidget().Rect
	x, y := uiinput.CursorPosition()
	layer := e.GetWidget().EffectiveInputLayer()
	if image.Pt(x, y).In(r) && uiinput.MouseButtonJustPressedLayer(ebiten.MouseButtonLeft, layer) {
		e.game.ui.SetFocusedWidget(e.TextInput)
		e.caret = e.pointPos(x, y)
		e.anchor = e.caret
		e.drag = true
		e.syncIME()
	}
	if e.drag {
		if y < r.Min.Y {
			e.top = max(0, e.top-1)
		} else if y >= r.Max.Y {
			e.top = min(len(e.starts)-1, e.top+1)
		}
		e.caret = e.pointPos(x, y)
		e.syncIME()
		e.drag = uiinput.MouseButtonPressed(ebiten.MouseButtonLeft)
	}
	if image.Pt(x, y).In(r) {
		_, wheel := ebiten.Wheel()
		e.top = max(0, min(max(0, len(e.starts)-max(1, (r.Dy()-8)/e.height)), e.top-int(wheel*3)))
	}
	if !e.IsFocused() || e.game.projects.pending != nil {
		e.ime.Cancel()
		return
	}
	if e.panel.editing {
		handled, _ := e.ime.Update()
		if handled {
			return
		}
	}
	e.keys()
}
func (e *documentSurface) setupIME() {
	e.ime.OnNewSession = func() *textinput.SessionOptions {
		a, b := min(e.anchor, e.caret), max(e.anchor, e.caret)
		e.imeBegin = e.starts[e.row(a)]
		e.imeEnd = e.starts[e.row(b)] + len(e.line(e.row(b)))
		rect := e.GetWidget().Rect.Inset(4)
		x := rect.Min.X + e.width(e.body[e.starts[e.row(e.caret)]:e.caret]) - e.left
		y := rect.Min.Y + (e.row(e.caret)-e.top)*e.height
		return &textinput.SessionOptions{CaretBounds: image.Rect(x, y, x+2, y+e.height), TextBeforeCaret: e.body[e.imeBegin:a], TextAfterCaret: e.body[b:e.imeEnd]}
	}
	e.ime.OnComposition = func(c *textinput.Composition) {
		e.composing = c.Text()
	}
	e.ime.OnCommit = func(c *textinput.Commit) {
		e.committing = true
		before, after := c.IsSurroundingTextReplaced()
		if before || after {
			prefix, suffix := c.SurroundingText()
			e.anchor = e.imeBegin
			e.caret = e.imeEnd
			e.insert(prefix + c.Text() + suffix)
			e.caret = min(len(e.body), e.imeBegin+len(prefix)+len(c.Text()))
			e.anchor = e.caret
		} else {
			e.insert(c.Text())
		}
		e.committing = false
	}
	e.ime.OnEndByUser = func() {
		e.game.ui.ClearFocus()
	}
}
func (e *documentSurface) keys() {
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)
	press := inpututil.IsKeyJustPressed
	if ctrl {
		switch {
		case press(ebiten.KeyA):
			e.anchor = 0
			e.caret = len(e.body)
		case press(ebiten.KeyC):
			e.game.copyDocument(e.body[min(e.anchor, e.caret):max(e.anchor, e.caret)])
		case press(ebiten.KeyX) && e.panel.editing:
			e.game.copyDocument(e.body[min(e.anchor, e.caret):max(e.anchor, e.caret)])
			e.insert("")
		case press(ebiten.KeyV) && e.panel.editing:
			e.insert(e.game.pasteDocument())
		case press(ebiten.KeyZ) && e.panel.editing && !shift:
			e.history(&e.undo, &e.redo)
		case (press(ebiten.KeyY) || press(ebiten.KeyZ) && shift) && e.panel.editing:
			e.history(&e.redo, &e.undo)
		default:
			return
		}
		e.syncIME()
		return
	}
	_, back := utf8.DecodeLastRuneInString(e.body[:e.caret])
	_, forward := utf8.DecodeRuneInString(e.body[e.caret:])
	row := e.row(e.caret)
	col := utf8.RuneCountInString(e.body[e.starts[row]:e.caret])
	repeat := func(key ebiten.Key) bool {
		return navigationRepeat(inpututil.KeyPressDuration(key))
	}
	switch {
	case repeat(ebiten.KeyArrowLeft):
		e.caret -= back
	case repeat(ebiten.KeyArrowRight):
		e.caret += forward
	case repeat(ebiten.KeyArrowUp), repeat(ebiten.KeyArrowDown):
		if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
			row = max(0, row-1)
		} else {
			row = min(len(e.starts)-1, row+1)
		}
		line := []rune(e.line(row))
		e.caret = e.starts[row] + len(string(line[:min(col, len(line))]))
	case press(ebiten.KeyHome):
		e.caret = e.starts[row]
	case press(ebiten.KeyEnd):
		e.caret = e.starts[row] + len(e.line(row))
	case e.panel.editing && repeat(ebiten.KeyBackspace):
		if e.anchor == e.caret {
			e.anchor -= back
		}
		e.insert("")
		return
	case e.panel.editing && repeat(ebiten.KeyDelete):
		if e.anchor == e.caret {
			e.anchor += forward
		}
		e.insert("")
		return
	case e.panel.editing && press(ebiten.KeyEnter):
		e.insert("\n")
		return
	default:
		chars := ebiten.AppendInputChars(nil)
		if len(chars) > 0 {
			e.insert(string(chars))
		}
		return
	}
	if !shift {
		e.anchor = e.caret
	}
	e.syncIME()
	e.reveal()
}
