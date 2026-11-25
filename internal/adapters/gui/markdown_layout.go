//go:build gui

// Measures and paints native rich-text runs in chat coordinates.
// Shared font sources keep headings and nested emphasis inexpensive across bubbles.
package gui

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gobolditalic"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/gomono"
)

type mdStyle struct {
	bold, italic, mono, strike, link bool
	size                             float64
}

var mdSources = sync.OnceValue(func() []*text.GoTextFaceSource {
	var sources []*text.GoTextFaceSource
	for _, data := range [][]byte{dejavuSansTTF, gobold.TTF, goitalic.TTF, gobolditalic.TTF, gomono.TTF} {
		src, err := text.NewGoTextFaceSource(bytes.NewReader(data))
		if err != nil {
			panic(err)
		}
		sources = append(sources, src)
	}
	return sources
})

func (s mdStyle) face() text.Face {
	i := 0
	if s.bold {
		i++
	}
	if s.italic {
		i += 2
	}
	if s.mono {
		i = 4
	}
	return &text.GoTextFace{Source: mdSources()[i], Size: s.size}
}

type mdRun struct {
	x, y, w, h float64
	style      mdStyle
	value      string
	fill       color.Color
}
type mdLayout struct {
	runs                              []mdRun
	left, right, x, y, height, extent float64
	space                             bool
	base                              float64
}

func (m *mdLayout) newline() {
	m.y += max(m.height, m.base*1.3)
	m.x = m.left
	m.height = 0
	m.space = false
}
func (m *mdLayout) finish() {
	if m.height > 0 {
		m.newline()
	}
	m.space = false
}
func (m *mdLayout) rectangle(x, y, w, h float64, c color.Color) {
	m.runs = append(m.runs, mdRun{x: x, y: y, w: w, h: h, fill: c})
	m.extent = max(m.extent, x+w)
}
func (m *mdLayout) run(value string, s mdStyle, w float64) {
	m.runs = append(m.runs, mdRun{x: m.x, y: m.y, w: w, style: s, value: value})
	m.x += w
	m.height = max(m.height, s.size*1.35)
	m.extent = max(m.extent, m.x)
}
func (m *mdLayout) word(value string, s mdStyle) {
	face := s.face()
	width := text.Advance(value, face)
	gap := 0.0
	if m.space && m.x > m.left {
		gap = text.Advance(" ", face)
	}
	if m.x > m.left && m.x+gap+width > m.right {
		m.newline()
	} else {
		m.x += gap
	}
	m.space = false
	if width <= m.right-m.x {
		m.run(value, s, width)
		return
	}
	// Break oversized tokens in linear time without splitting a UTF-8 encoding.
	for len(value) > 0 {
		r, n := utf8.DecodeRuneInString(value)
		w := text.Advance(string(r), face)
		if m.x > m.left && m.x+w > m.right {
			m.newline()
		}
		m.run(value[:n], s, w)
		value = value[n:]
	}
}
func (m *mdLayout) prose(value string, s mdStyle) {
	for value != "" {
		n := strings.IndexAny(value, " \n\r\t")
		if n < 0 {
			m.word(value, s)
			return
		}
		if n > 0 {
			m.word(value[:n], s)
		}
		m.space = true
		value = value[n+1:]
	}
}
func (m *mdLayout) paint(dst *ebiten.Image, r image.Rectangle) {
	for _, run := range m.runs {
		x, y := float64(r.Min.X)+run.x, float64(r.Min.Y)+run.y
		if y > float64(dst.Bounds().Max.Y) || y+max(run.h, run.style.size*2) < float64(dst.Bounds().Min.Y) {
			continue
		}
		if run.fill != nil {
			vector.FillRect(dst, float32(x), float32(y), float32(run.w), float32(run.h), run.fill, false)
			continue
		}
		c := color.Color(color.White)
		if run.style.link {
			c = color.RGBA{166, 208, 255, 255}
		}
		op := &text.DrawOptions{}
		op.GeoM.Translate(x, y)
		op.ColorScale.ScaleWithColor(c)
		text.Draw(dst, run.value, run.style.face(), op)
		if run.style.strike || run.style.link {
			offset := run.style.size * 1.12
			if run.style.strike {
				offset = run.style.size * .62
			}
			vector.FillRect(dst, float32(x), float32(y+offset), float32(run.w), 1, c, false)
		}
	}
}
func (m *mdLayout) size() (int, int) {
	return max(1, int(math.Ceil(m.extent))), max(1, int(math.Ceil(m.y+m.height)))
}

type markdownButton struct {
	*widget.Button
	layout *mdLayout
}

func (b *markdownButton) PreferredSize() (int, int) {
	return b.layout.size()
}
func (b *markdownButton) Render(dst *ebiten.Image) {
	b.Button.Render(dst)
	b.layout.paint(dst, b.GetWidget().Rect)
}

type markdownPreview struct {
	*widget.Text
	layout      *mdLayout
	width, font int
}

func (p *markdownPreview) PreferredSize() (int, int) {
	return p.layout.size()
}
func (p *markdownPreview) Render(dst *ebiten.Image) {
	p.layout.paint(dst, p.GetWidget().Rect)
}
func (p *markdownPreview) set(value string) {
	if p.Label != value || p.layout == nil {
		p.Label = value
		p.layout = layoutMarkdown(value, p.width, p.font)
	}
}

func (g *Game) chatBubbleWidth() int {
	pane := g.hudPreferredWidth()
	if g.inWorkspace() {
		pane = g.detailPanePreferredWidth()
	}
	return max(60, min(540, pane*4/5-24))
}
func (g *Game) markdownText(value string, width int) *markdownPreview {
	p := &markdownPreview{Text: widget.NewText(widget.TextOpts.Text("", g.theme.ButtonTheme.TextFace, g.theme.ButtonTheme.TextColor.Idle)), width: width, font: g.prefs.Font}
	p.set(value)
	return p
}
