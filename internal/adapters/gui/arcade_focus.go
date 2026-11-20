//go:build gui

// Lends the embedded gallery and its modal settings the HUD navigation cursor.
// Activation uses the same surface actions and geometry as pointer input.
package gui

import (
	"image"

	"ebe-boilerplate/surface"
	"github.com/hajimehoshi/ebiten/v2"
)

const galleryToolbarItems = 6

func (g *Game) galleryInput(d image.Point, back, activate, tab bool) bool {
	if g.nav.owner != "stmpck" {
		return false
	}
	s := g.arcade.session.Load()
	if s == nil {
		return false
	}
	if s.Settings {
		g.nav.active = true
		if back {
			s.Back()
			return true
		}
		g.nav.setting = max(0, min(4, g.nav.setting+d.Y))
		if activate {
			r := g.settingRect()
			s.ActivateMenu(float64(r.Min.X+4), float64(r.Min.Y+4))
		}
		return true
	}
	if !g.nav.gallery {
		if tab {
			g.enterBody()
			return true
		}
		return false
	}
	if back || tab {
		g.focusStrip()
		return true
	}
	item := g.nav.galleryItem
	total := len(s.Order())
	if item >= galleryToolbarItems && total == 0 {
		item = 0
	}
	if item < galleryToolbarItems {
		if d == g.inward().Mul(-1) && (d.Y < 0 || d.X < 0 && item == 0 || d.X > 0 && item == galleryToolbarItems-1) {
			g.focusStrip()
			return true
		}
		if d.Y > 0 && total > 0 {
			item = galleryToolbarItems + s.Focus
		} else {
			item = max(0, min(galleryToolbarItems-1, item+d.X))
		}
	} else {
		if d.Y < 0 && s.Focus < 3 {
			item = 1 + s.Tab
		} else if d == g.inward().Mul(-1) && (d.Y > 0 && s.Focus+3 >= total || d.X < 0 && s.Focus%3 == 0 || d.X > 0 && (s.Focus%3 == 2 || s.Focus == total-1)) {
			g.focusStrip()
			return true
		} else {
			offset := d.X + 3*d.Y
			if d.X != 0 && (s.Focus%3+d.X < 0 || s.Focus%3+d.X >= 3) {
				offset = 0
			}
			s.Move(offset, g.arcade.canvas.Bounds().Dy())
			item = galleryToolbarItems + s.Focus
		}
	}
	g.nav.galleryItem = item
	if activate {
		if item >= galleryToolbarItems {
			s.Open(s.Focus)
		} else {
			r := g.galleryRect(s)
			s.ActivateMenu(float64((r.Min.X+r.Max.X)/2), float64((r.Min.Y+r.Max.Y)/2))
			g.nav.setting = 0
		}
	}
	return true
}
func (g *Game) settingRect() image.Rectangle {
	i := g.nav.setting
	if i == 0 {
		return image.Rect(8, 4, 56, 48)
	}
	y := 100 + (i-1)*58
	return image.Rect(16, y, g.arcade.canvas.Bounds().Dx()-16, y+50)
}
func (g *Game) galleryRect(s *surface.Game) image.Rectangle {
	w := g.arcade.canvas.Bounds().Dx()
	i := g.nav.galleryItem
	if i >= galleryToolbarItems {
		x, y, width, height := s.TileRect(s.Focus, w)
		return image.Rect(int(x), int(y), int(x+width), int(y+height))
	}
	return s.ToolbarRect(i)
}
func (g *Game) drawArcadeFocus(dst *ebiten.Image) {
	s := g.arcade.session.Load()
	if !g.nav.active || s == nil {
		return
	}
	var r image.Rectangle
	switch {
	case s.Settings:
		r = g.settingRect()
	case s.Selected >= 0:
		return
	case g.nav.gallery:
		r = g.galleryRect(s)
	default:
		g.drawFocus(dst)
		return
	}
	scale := s.InputScale
	r = image.Rect(int(float64(r.Min.X)*scale), int(float64(r.Min.Y)*scale), int(float64(r.Max.X)*scale), int(float64(r.Max.Y)*scale))
	drawFocusRect(dst, r.Add(image.Pt(s.OffsetX, s.OffsetY)))
}
