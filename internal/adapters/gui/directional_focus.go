//go:build gui

// Resolves directional focus against current tab widgets and their laid-out bounds.
// Remembers destinations across rebuilds and reveals the owning viewport.
package gui

import (
	"image"
	"image/color"
	"math"
	"slices"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type focusTarget struct {
	w    widget.Focuser
	pane *widget.ScrollContainer
}
type focusBookmark struct {
	index int
	label string
}
type navigation struct {
	active, strip, gallery bool
	resizing               bool
	galleryItem, setting   int
	target                 focusTarget
	memory                 map[hud.TabID]focusBookmark
	body                   widget.PreferredSizeLocateableWidget
	owner                  hud.TabID
}

func collectTargets(node widget.PreferredSizeLocateableWidget, targets *[]focusTarget) {
	if node.GetWidget().Disabled || !node.GetWidget().IsVisible() {
		return
	}
	switch w := node.(type) {
	case *boundedPreferredWidget:
		collectTargets(w.child, targets)
	case *widget.Container:
		for _, child := range w.Children() {
			collectTargets(child, targets)
		}
	case *widget.ScrollContainer:
		for _, f := range w.GetFocusers() {
			*targets = append(*targets, focusTarget{f, w})
		}
	case widget.Focuser:
		if w.TabOrder() >= 0 {
			*targets = append(*targets, focusTarget{w, nil})
		}
	}
}
func (g *Game) bodyTargets() []focusTarget {
	var targets []focusTarget
	if g.nav.owner == hud.TabCluster && g.cluster.gear != nil {
		collectTargets(g.cluster.gear, &targets)
	}
	collectTargets(g.nav.body, &targets)
	return targets
}
func focusLabel(f widget.Focuser) string {
	w, ok := f.(interface{ Text() *widget.Text })
	if ok && w.Text() != nil {
		return w.Text().Label
	}
	return ""
}
func (g *Game) focusBody(targets []focusTarget, i int) {
	if len(targets) == 0 {
		g.focusStrip()
		return
	}
	g.ui.ClearFocus()
	g.nav.active = true
	g.nav.strip = false
	g.nav.target = targets[i]
	if targets[i].w.GetWidget() != g.divider {
		g.nav.resizing = false
	}
	if g.nav.memory == nil {
		g.nav.memory = make(map[hud.TabID]focusBookmark)
	}
	if g.actions.mode != "" {
		g.actions.focusIndex = i
	} else {
		g.nav.memory[g.nav.owner] = focusBookmark{i, focusLabel(targets[i].w)}
	}
	t := targets[i]
	if t.pane != nil {
		r := t.w.GetWidget().Rect
		v := t.pane.ViewRect()
		moveBodyScroll(t.pane, float64(max(0, v.Min.Y-r.Min.Y)-max(0, r.Max.Y-v.Max.Y)))
	}
}
func (g *Game) enterBody() {
	if g.actions.mode != "" {
		targets := g.bodyTargets()
		g.focusBody(targets, min(g.actions.focusIndex, max(0, len(targets)-1)))
		return
	}
	if g.nav.owner == "stmpck" {
		g.nav.active = true
		g.nav.gallery = true
		g.nav.strip = false
		g.ui.ClearFocus()
		return
	}
	targets := g.bodyTargets()
	m := g.nav.memory[g.nav.owner]
	i := min(m.index, max(0, len(targets)-1))
	if m.label != "" && len(targets) > 0 && focusLabel(targets[i].w) != m.label {
		for j, t := range targets {
			if focusLabel(t.w) == m.label {
				i = j
				break
			}
		}
	}
	g.focusBody(targets, i)
}
func (g *Game) focusStrip() {
	if g.actions.mode != "" {
		g.enterBody()
		return
	}
	g.ui.ClearFocus()
	g.nav.active = true
	g.nav.strip = true
	g.nav.gallery = false
	g.nav.resizing = false
	g.nav.target = focusTarget{g.tabButtons[g.shell.Snapshot().ActiveIndex], g.tabScroll}
	g.requestActiveTabVisible()
}
func (g *Game) ownsFocus(f widget.Focuser) bool {
	b, ok := f.(*widget.Button)
	if ok && slices.Contains(g.tabButtons, b) {
		return true
	}
	return slices.ContainsFunc(g.bodyTargets(), func(t focusTarget) bool {
		return t.w == f
	})
}
func (g *Game) setNavTarget(f widget.Focuser) {
	targets := g.bodyTargets()
	for i, t := range targets {
		if t.w == f {
			g.focusBody(targets, i)
			return
		}
	}
	g.focusStrip()
}
func (g *Game) inward() image.Point {
	switch g.sideRail() {
	case "left_rail":
		return image.Pt(1, 0)
	case "right_rail":
		return image.Pt(-1, 0)
	}
	if g.bottomTabs() {
		return image.Pt(0, -1)
	}
	return image.Pt(0, 1)
}
func (g *Game) moveFocus(d image.Point) {
	if !g.nav.active || g.nav.target.w == nil {
		g.focusStrip()
	}
	in := g.inward()
	if g.nav.strip {
		if d == in {
			g.enterBody()
			return
		}
		if d != in.Mul(-1) {
			s := g.shell.Snapshot()
			g.selectTabWithReveal((s.ActiveIndex + d.X + d.Y + len(s.Tabs)) % len(s.Tabs))
		}
		return
	}
	targets := g.bodyTargets()
	r := g.nav.target.w.GetWidget().Rect
	best := math.Inf(1)
	index := -1
	for i, t := range targets {
		bounds := t.w.GetWidget().Rect
		if t.pane != nil && t.pane != g.nav.target.pane {
			bounds = bounds.Intersect(t.pane.ViewRect())
		}
		score := directionScore(r, bounds, d)
		if score < best {
			best = score
			index = i
		}
	}
	if index >= 0 {
		g.focusBody(targets, index)
	} else if d == in.Mul(-1) {
		g.focusStrip()
	}
}

// Require forward separation on the movement axis. Overlapping perpendicular
// spans form a row or column and take priority over diagonal candidates.
func directionScore(a, b image.Rectangle, d image.Point) float64 {
	if a.Empty() || b.Empty() {
		return math.Inf(1)
	}
	if d.Y != 0 {
		a = image.Rect(a.Min.Y, a.Min.X, a.Max.Y, a.Max.X)
		b = image.Rect(b.Min.Y, b.Min.X, b.Max.Y, b.Max.X)
	}
	if d.X+d.Y < 0 {
		a, b = b, a
	}
	if b.Min.X < a.Max.X-1 {
		return math.Inf(1)
	}
	gap := max(0, max(a.Min.Y-b.Max.Y, b.Min.Y-a.Max.Y))
	offset := math.Abs(float64((a.Min.Y+a.Max.Y)-(b.Min.Y+b.Max.Y))) / 2
	score := float64(b.Min.X-a.Max.X) + offset*0.1
	if gap > 0 {
		score += 1e7 + float64(gap)*4
	}
	return score
}
func (g *Game) stepFocus(step int) {
	ts := g.bodyTargets()
	i := slices.IndexFunc(ts, func(t focusTarget) bool {
		return t.w == g.nav.target.w
	})
	if i < 0 && step < 0 {
		i = len(ts)
	}
	i += step
	if i < 0 || i >= len(ts) {
		if g.actions.mode != "" && len(ts) > 0 {
			g.focusBody(ts, (i+len(ts))%len(ts))
			return
		}
		g.focusStrip()
		return
	}
	g.focusBody(ts, i)
}
func (g *Game) activateFocus() {
	if !g.nav.active || g.nav.strip {
		g.enterBody()
		return
	}
	if g.nav.target.w == nil || !g.ownsFocus(g.nav.target.w) {
		return
	}
	if g.nav.target.w.GetWidget() == g.divider {
		g.nav.resizing = true
		return
	}
	switch w := g.nav.target.w.(type) {
	case *widget.TextInput:
		g.ui.SetFocusedWidget(w)
	case interface{ Click() }:
		w.Click()
	}
}
func (g *Game) drawFocus(dst *ebiten.Image) {
	g.applyProjectFocus()
	if g.nav.active && !g.nav.strip && !g.nav.gallery && g.nav.target.w == nil {
		g.enterBody()
	}
	if !g.nav.active || g.nav.gallery || g.nav.target.w == nil {
		return
	}
	r := g.nav.target.w.GetWidget().Rect
	p := g.nav.target.pane
	if p != nil {
		r = r.Intersect(p.ViewRect())
	}
	drawFocusRect(dst, r)
}
func drawFocusRect(dst *ebiten.Image, r image.Rectangle) {
	r = r.Inset(2)
	if !r.Empty() {
		vector.StrokeRect(dst, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), 3, color.RGBA{255, 216, 96, 255}, false)
	}
}

// An editor's parent changes on resize or refresh. Keep its native selection,
// caret and draft only while the same field in the same view is being rebuilt.
type editorIdentity struct {
	tab                          hud.TabID
	section                      int
	project, mode, thread, label string
	theme                        *widget.Theme
}

func (g *Game) editorForLayout(label string, create func() *widget.TextInput) *widget.TextInput {
	tab := g.shell.Snapshot().ActiveTab().ID()
	w := &g.projects.workspace
	key := editorIdentity{tab, g.selectedSections[tab], w.id, w.mode, w.thread, label, g.theme}
	if g.ui != nil {
		old, ok := g.ui.GetFocusedWidget().(*widget.TextInput)
		if ok && old.GetWidget().CustomData == key {
			reattachEditor(old)
			return old
		}
	}
	field := create()
	field.GetWidget().CustomData = key
	return field
}

func (g *Game) restoreEditing(old widget.Focuser) {
	field, ok := old.(*widget.TextInput)
	if !ok || g.ui.GetFocusedWidget() != nil {
		return
	}
	for _, target := range g.bodyTargets() {
		if target.w == field {
			g.ui.SetFocusedWidget(field)
			g.nav.target = target
			return
		}
	}
}
