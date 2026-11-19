//go:build gui

// Composes the private game surface in Apparat's safe viewport.
// Owns lazy initialization, game texture and local pointer transform.
package gui

import (
	"ebe-boilerplate/surface"
	"github.com/hajimehoshi/ebiten/v2"
	"path/filepath"
	"sync/atomic"
)

type arcadeView struct {
	session atomic.Pointer[surface.Game]
	canvas  *ebiten.Image
	active  bool
	back    atomic.Bool
}

func (g *Game) arcadeUpdate() bool {
	a := &g.arcade
	s := a.session.Load()
	if g.shell.Snapshot().ActiveTab().ID() != "stmpck" {
		if a.active {
			s.Suspend()
			a.active = false
		}
		return false
	}
	if s == nil {
		s = surface.New(filepath.Join(g.runtimeInfo.RuntimePath, "stmpck"))
		a.session.Store(s)
	}
	a.active = true
	if a.back.Swap(false) {
		s.RequestBack()
	}
	top := 0
	gallery := s.Selected < 0 && !s.Settings
	if gallery {
		top = 116
	}
	width := g.width
	s.OffsetX = 0
	if gallery && g.sideRail() != "" {
		top = 60
		width = g.contentWidth() - windowMargin
		if g.sideRail() == "left_rail" {
			s.OffsetX = railGutter + windowMargin
		}
	}
	bottom := 0
	if gallery && g.bottomTabs() {
		top = 60
		bottom = diagnosticsHeight + tabHeight + bodyGap
	}
	w, h := s.Layout(width, max(1, g.height-top-bottom))
	s.OffsetY = top
	s.InputScale = float64(width) / float64(w)
	if a.canvas == nil || a.canvas.Bounds().Dx() != w || a.canvas.Bounds().Dy() != h {
		if a.canvas != nil {
			a.canvas.Deallocate()
		}
		a.canvas = ebiten.NewImage(w, h)
	}
	_ = s.Update()
	g.activeTabID.Store("stmpck")
	return !gallery
}
func (g *Game) arcadeDraw(dst *ebiten.Image) bool {
	a := &g.arcade
	if g.shell.Snapshot().ActiveTab().ID() != "stmpck" || a.canvas == nil {
		return false
	}
	s := a.session.Load()
	s.Draw(a.canvas)
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(s.InputScale, s.InputScale)
	op.GeoM.Translate(float64(s.OffsetX), float64(s.OffsetY))
	dst.DrawImage(a.canvas, &op)
	return true
}
func (g *Game) ArcadeSuspend() {
	s := g.arcade.session.Load()
	if s != nil {
		s.Suspend()
	}
}
func (g *Game) ArcadeBack() {
	g.arcade.back.Store(true)
}
