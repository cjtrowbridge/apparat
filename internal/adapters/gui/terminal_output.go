//go:build gui

// Shares literal terminal rendering, pointer selection and clipboard controls.
// One visible session owns the frozen display until selection ends or its address changes.
package gui

import (
	"github.com/cjtrowbridge/apparat/internal/hud"
	"image"
	"math"
	"slices"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type terminalOutputID struct {
	chat                        modelAddress
	session, host, page, cursor string
}

type terminalOutput struct {
	game    *Game
	id      terminalOutputID
	passage selectablePassage
	source  string
	size    float64
	width   int
	live    *chatTerminalSurface
}

func terminalLineHeight(size float64) int {
	return int(math.Ceil(size * 1.35))
}

func (g *Game) detachMissingOutput() {
	o := g.terminalOutput
	if o != nil && !slices.Contains(g.bubbleText.views, &o.passage) {
		g.terminalOutput = nil
	}
}

func (g *Game) bindOutput(id terminalOutputID, w *widget.Widget, row *widget.Container, size float64, live *chatTerminalSurface) *terminalOutput {
	o := g.terminalOutput
	if o == nil || o.id != id {
		o = &terminalOutput{game: g, id: id}
		g.terminalOutput = o
	}
	p := &o.passage
	o.live = live
	p.view = w
	p.row = row
	p.inset = 4
	if p.layout == nil || size != o.size {
		o.width = -1
		o.size = size
	}
	p.tap = nil
	if live != nil {
		p.tap = func() {
			g.ui.SetFocusedWidget(live.TextInput)
		}
	}
	label := "Copy selection"
	if live != nil {
		label = "Copy"
	}
	p.copy = g.projectButton(row, label, o.copySelection)
	p.copy.GetWidget().Disabled = p.layout == nil || p.selected() == ""
	if !p.copy.GetWidget().Disabled {
		g.bubbleText.active = p
	}
	g.bubbleText.views = append(g.bubbleText.views, p)
	return o
}

func (o *terminalOutput) frozen() bool {
	p := &o.passage
	return o.game.bubbleText.pending == p || p.layout != nil && p.layout.selectionStart != p.layout.selectionEnd
}
func (o *terminalOutput) copySelection() {
	if o.game.terminalOutput == o && o.passage.layout != nil {
		o.game.copyDocument(o.passage.selected())
	}
}
func (e *chatTerminalSurface) paste() {
	if e.output == nil || e.g.terminalOutput != e.output || e.id != e.t.selected || e.t.permissionPage {
		return
	}
	e.g.bubbleText.clear()
	e.g.pasteInto(e.enqueue)
}

func (o *terminalOutput) update(value string, width int) {
	width = max(1, width-8)
	p := &o.passage
	if p.layout != nil && o.frozen() {
		value = o.source
	}
	if p.layout != nil && value == o.source && width == o.width {
		return
	}
	old := p.layout
	o.source, o.width = value, width
	p.layout = literalTerminalLayout(value, width, o.size, o.live == nil)
	if old != nil && o.game.bubbleText.active == p {
		p.layout.selectionStart = old.selectionStart
		p.layout.selectionEnd = old.selectionEnd
	}
}

// Newlines and tabs are exact clipboard bytes. Only the geometry wraps.
func literalTerminalLayout(value string, width int, size float64, wrap bool) *mdLayout {
	style := mdStyle{mono: true, size: size}
	face := style.face()
	cell := text.Advance("M", face)
	lineHeight := float64(terminalLineHeight(size))
	m := &mdLayout{right: float64(width), base: size, terminalRows: []int{0}}
	start := 0
	advance := 0.0
	flush := func(end int) {
		if end > start {
			m.run(value[start:end], style, advance)
		}
		advance = 0
		start = end
	}
	nextLine := func() {
		m.terminalRows = append(m.terminalRows, m.plain.Len())
		m.y += lineHeight
		m.x = 0
		m.height = 0
	}
	for i, r := range value {
		if r == '\n' {
			flush(i)
			m.plain.WriteByte('\n')
			start = i + 1
			nextLine()
			continue
		}
		w := text.Advance(string(r), face)
		if r == '\t' {
			flush(i)
			next := (math.Floor(m.x/(8*cell)) + 1) * 8 * cell
			m.run("\t", style, next-m.x)
			start = i + 1
			continue
		}
		if wrap && m.x+advance > 0 && m.x+advance+w > float64(width) {
			flush(i)
			nextLine()
		}
		advance += w
	}
	flush(len(value))
	m.height = lineHeight
	return m
}

type terminalTranscript struct {
	*widget.Text
	output *terminalOutput
	value  string
}

func (v *terminalTranscript) PreferredSize() (int, int) {
	_, h := v.output.passage.layout.size()
	return 44, h + 8
}
func (v *terminalTranscript) SetLocation(r image.Rectangle) {
	v.Text.SetLocation(r)
	v.output.update(v.value, r.Dx())
}
func (v *terminalTranscript) Render(dst *ebiten.Image) {
	r := v.GetWidget().Rect
	clip := r.Intersect(dst.Bounds())
	if !clip.Empty() {
		v.output.passage.layout.paint(dst.SubImage(clip).(*ebiten.Image), r.Inset(4))
	}
}

func (g *Game) terminalTranscript(c *widget.Container, id terminalOutputID, value string) {
	row := controlRow(c)
	v := &terminalTranscript{Text: g.detailText(""), value: value}
	v.GetWidget().LayoutData = widget.RowLayoutData{Stretch: true}
	v.output = g.bindOutput(id, v.GetWidget(), row, float64(max(10, min(16, g.prefs.Font))), nil)
	v.output.update(value, g.detailPanePreferredWidth())
	c.AddChild(v)
}

func (g *Game) terminalClipboardActions() []catalogAction {
	o := g.terminalOutput
	if o == nil || o.passage.layout == nil {
		return nil
	}
	if o.live != nil && (o.live.t.permissionPage || o.live.t.selected != o.id.session) {
		return nil
	}
	owner := o.id.chat.project + "/" + o.id.chat.session + "/" + o.id.host + "/" + o.id.session + "/" + o.id.page + "/" + o.id.cursor
	action := func(id, name string, run func()) catalogAction {
		address := hud.ActionAddress{Operation: "task", Kind: "terminal-clipboard", Owner: owner, Target: id}
		return catalogAction{hud.Destination{Address: address, Name: name, Path: "Terminal / " + name}, run}
	}
	var actions []catalogAction
	if o.passage.selected() != "" {
		actions = append(actions, action("copy", "Copy selection", o.copySelection))
	}
	if o.live != nil {
		actions = append(actions, action("paste", "Paste", o.live.paste))
	}
	return actions
}
