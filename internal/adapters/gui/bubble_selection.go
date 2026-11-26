//go:build gui

// Gives a chat passage exclusive ownership of mouse drags or a stationary touch hold.
// A moving touch transfers to the existing scroll pane before selecting any text.
package gui

import (
	"image"
	"image/color"
	"time"

	uiinput "github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var selectionColor = color.RGBA{R: 92, G: 95, B: 174, A: 255}

type selectableBubble struct {
	view  *markdownButton
	tools *bubbleTools
	copy  *widget.Button
}

func (p *selectableBubble) selected() string {
	m := p.view.layout
	return m.plain.String()[m.selectionStart:m.selectionEnd]
}
func (p *selectableBubble) position(x, y int) int {
	r := p.view.GetWidget().Rect
	return p.view.layout.position(float64(x-r.Min.X), float64(y-r.Min.Y))
}

func (p *selectableBubble) showCopy(show bool) {
	visible := widget.Visibility_Hide
	if show {
		visible = widget.Visibility_Show
	}
	if p.copy.GetWidget().GetVisibility() != visible {
		p.copy.GetWidget().SetVisibility(visible)
		p.tools.row.RequestRelayout()
	}
}

type bubbleTextSelection struct {
	views                      []*selectableBubble
	pending, active            *selectableBubble
	start                      image.Point
	anchor, wordStart, wordEnd int
	held                       time.Time
	dragging                   bool
}

func (s *bubbleTextSelection) reset() {
	*s = bubbleTextSelection{}
}
func (s *bubbleTextSelection) clear() {
	if s.active != nil {
		s.active.view.layout.selectionStart = 0
		s.active.view.layout.selectionEnd = 0
		s.active.showCopy(false)
		s.active = nil
	}
}
func (g *Game) beginBubble(x, y int, touch bool, id ebiten.TouchID) bool {
	if g.actions.mode != "" || g.gesture.Owner != gestureNone || g.pointInPTT(x, y) || g.pointInDebugOverlay(x, y) {
		return false
	}
	pt := image.Pt(x, y)
	s := &g.bubbleText
	p := s.active
	if p != nil && pt.In(p.tools.row.GetWidget().Rect) {
		return false
	}
	s.clear()
	if g.verticalScrollAt(x, y) == nil {
		return false
	}
	for _, p := range s.views {
		if !pt.In(p.view.GetWidget().Rect) || !p.view.GetWidget().EffectiveInputLayer().ActiveFor(x, y, uiinput.LayerEventTypeMouseButton) {
			continue
		}
		if !g.claimGesture(gestureBubble, touch, id) {
			return false
		}
		s.pending = p
		s.start = pt
		s.anchor = p.position(x, y)
		s.held = time.Now()
		s.dragging = false
		return true
	}
	return false
}
func (g *Game) selectPassage(a, b int) {
	s := &g.bubbleText
	p := s.pending
	m := p.view.layout
	m.selectionStart, m.selectionEnd = min(a, b), max(a, b)
	if !s.dragging {
		g.ui.ClearFocus()
		g.planning.bubbles.show(p.tools, true)
	}
	s.active = p
	s.dragging = true
	p.showCopy(a != b)
}
func (g *Game) finishBubble() {
	if g.bubbleText.dragging {
		g.bodyScrollCancelUpdates = bodyScrollCancelUpdateFrames
	}
	g.bubbleText.pending = nil
	g.releaseGesture(gestureBubble)
}
func (g *Game) bubbleMouse(x, y int) bool {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.beginBubble(x, y, false, 0)
	}
	if g.gesture.Owner != gestureBubble || g.gesture.Touch {
		return false
	}
	s := &g.bubbleText
	if s.pending == nil || inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		g.finishBubble()
		return true
	}
	pos := s.pending.position(x, y)
	if pos != s.anchor || s.dragging {
		g.selectPassage(s.anchor, pos)
	}
	return true
}
func (g *Game) updateBubbleTouch() {
	s := &g.bubbleText
	id := g.gesture.ID
	if s.pending == nil || inpututil.IsTouchJustReleased(id) {
		g.finishBubble()
		return
	}
	x, y := ebiten.TouchPosition(id)
	dx, dy := x-s.start.X, y-s.start.Y
	if !s.dragging && dx*dx+dy*dy > 100 {
		g.finishBubble()
		if g.beginBodyTouch(id, s.start.X, s.start.Y) {
			g.updateBodyTouch()
		}
		return
	}
	if !s.dragging && time.Since(s.held) >= 450*time.Millisecond {
		s.wordStart, s.wordEnd = wordAt(s.pending.view.layout.plain.String(), s.anchor)
		g.selectPassage(s.wordStart, s.wordEnd)
		return
	}
	if s.dragging && dx*dx+dy*dy > 16 {
		pos := s.pending.position(x, y)
		g.selectPassage(min(s.wordStart, pos), max(s.wordEnd, pos))
	}
}
