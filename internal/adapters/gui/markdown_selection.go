//go:build gui

// Maps laid-out Markdown to byte ranges in its displayed plain text.
// Soft wrapping changes coordinates only. Block and hard breaks enter the text.
package gui

import (
	"math"
	"sort"
	"unicode"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (m *mdLayout) hardBreak() {
	m.plain.WriteByte('\n')
	m.newline()
}

func (m *mdLayout) highlight(dst *ebiten.Image, r mdRun, x, y float64) {
	a, b := max(m.selectionStart-r.start, 0), min(m.selectionEnd-r.start, len(r.value))
	if a >= b {
		return
	}
	face := r.style.face()
	left := text.Advance(r.value[:a], face)
	right := text.Advance(r.value[:b], face)
	if m.terminalRows != nil && r.value == "\t" {
		right = r.w
	}
	vector.FillRect(dst, float32(x+left), float32(y), float32(right-left), float32(r.style.size*1.35), selectionColor, false)
}

func (m *mdLayout) position(x, y float64) int {
	rows := m.terminalRows
	if len(rows) > 0 {
		height := terminalLineHeight(m.base)
		row := min(max(0, int(y)/height), len(rows)-1)
		top := float64(row * height)
		idx := sort.Search(len(m.runs), func(i int) bool {
			return m.runs[i].y >= top
		})
		pos := rows[row]
		for _, r := range m.runs[idx:] {
			if r.y != top {
				break
			}
			if x < r.x+r.w {
				return runPosition(&r, x)
			}
			pos = r.start + len(r.value)
		}
		return pos
	}
	var best *mdRun
	nearest := math.Inf(1)
	for i := range m.runs {
		r := &m.runs[i]
		if r.fill != nil || r.value == "" {
			continue
		}
		dx := max(r.x-x, 0, x-r.x-r.w)
		dy := max(r.y-y, 0, y-r.y-r.style.size*1.35)
		distance := math.Hypot(dx, dy*3)
		if distance < nearest {
			best, nearest = r, distance
		}
	}
	if best == nil {
		return 0
	}
	return runPosition(best, x)
}

func runPosition(best *mdRun, x float64) int {
	if best.value == "\t" {
		if x < best.x+best.w/2 {
			return best.start
		}
		return best.start + 1
	}
	if x <= best.x {
		return best.start
	}
	if x >= best.x+best.w {
		return best.start + len(best.value)
	}
	chars := []rune(best.value)
	face := best.style.face()
	i := sort.Search(len(chars), func(i int) bool {
		before := text.Advance(string(chars[:i]), face)
		after := text.Advance(string(chars[:i+1]), face)
		return x-best.x < (before+after)/2
	})
	return best.start + len(string(chars[:i]))
}

func wordAt(s string, i int) (int, int) {
	if i == len(s) && i > 0 {
		_, n := utf8.DecodeLastRuneInString(s)
		i -= n
	}
	a, b := i, i
	for a > 0 {
		r, n := utf8.DecodeLastRuneInString(s[:a])
		if unicode.IsSpace(r) {
			break
		}
		a -= n
	}
	for b < len(s) {
		r, n := utf8.DecodeRuneInString(s[b:])
		if unicode.IsSpace(r) {
			break
		}
		b += n
	}
	return a, b
}
