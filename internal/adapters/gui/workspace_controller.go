//go:build gui

// Routes standard controller navigation through existing Projects buttons and inputs.
// Focus changes reveal the selected control within its owning scrolling pane.
package gui

import (
	"math"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) projectPad(id ebiten.GamepadID) {
	if g.shell.Snapshot().ActiveTab().ID() != hud.TabProjects {
		return
	}
	dir := 0
	switch {
	case inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftTop):
		dir = -1
	case inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftBottom):
		dir = 1
	}
	activate := inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightBottom)
	if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightRight) {
		w := &g.projects.workspace
		if w.id != "" && w.mode != "menu" {
			w.mode = "menu"
			g.layoutDirty = true
		} else if w.id != "" {
			*w = workspacePanel{}
			g.layoutDirty = true
		} else if g.collapsed() {
			g.showList(hud.TabProjects)
		}
		return
	}
	if dir != 0 || activate {
		g.projectPadFocus(dir, activate)
	}
}
func (g *Game) projectPadFocus(dir int, activate bool) {
	type target struct {
		w      widget.Focuser
		scroll *widget.ScrollContainer
	}
	var targets []target
	current := -1
	for _, pane := range g.verticalScrolls {
		for _, f := range pane.GetFocusers() {
			if f.IsFocused() {
				current = len(targets)
			}
			targets = append(targets, target{f, pane})
		}
	}
	if len(targets) == 0 {
		return
	}
	next := 0
	if current >= 0 {
		next = (current + dir + len(targets)) % len(targets)
		if current != next {
			targets[current].w.Focus(false)
		}
	}
	t := targets[next]
	t.w.Focus(true)
	rect := t.w.GetWidget().Rect
	viewport := t.scroll.ViewRect()
	overflow := t.scroll.ContentRect().Dy() - viewport.Dy()
	if overflow > 0 {
		step := min(0, rect.Min.Y-viewport.Min.Y) + max(0, rect.Max.Y-viewport.Max.Y)
		t.scroll.ScrollTop = math.Max(0, math.Min(1, t.scroll.ScrollTop+float64(step)/float64(overflow)))
	}
	if activate && current >= 0 {
		b, ok := t.w.(*widget.Button)
		if ok {
			b.Click()
		}
	}
}
