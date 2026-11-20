//go:build gui

// Unifies keyboard and first-standard-controller input at the HUD boundary.
// Native text editing and active embedded games retain their own input.
package gui

import (
	"image"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var directions = [...]struct {
	delta  image.Point
	key    ebiten.Key
	button ebiten.StandardGamepadButton
}{
	{image.Pt(-1, 0), ebiten.KeyArrowLeft, ebiten.StandardGamepadButtonLeftLeft},
	{image.Pt(1, 0), ebiten.KeyArrowRight, ebiten.StandardGamepadButtonLeftRight},
	{image.Pt(0, -1), ebiten.KeyArrowUp, ebiten.StandardGamepadButtonLeftTop},
	{image.Pt(0, 1), ebiten.KeyArrowDown, ebiten.StandardGamepadButtonLeftBottom},
}

func navigationRepeat(frames int) bool {
	return frames == 1 || frames >= 24 && (frames-24)%6 == 0
}

func (g *Game) navigateInputs() {
	pad := ebiten.GamepadID(-1)
	for _, id := range ebiten.AppendGamepadIDs(nil) {
		if ebiten.IsStandardGamepadLayoutAvailable(id) {
			pad = id
			break
		}
	}
	pressed := func(b ebiten.StandardGamepadButton) bool {
		return pad >= 0 && inpututil.IsStandardGamepadButtonJustPressed(pad, b)
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || len(inpututil.AppendJustPressedTouchIDs(nil)) > 0 {
		g.nav.resizing = false
		g.nav.active = false
		g.nav.gallery = false
	}
	back := inpututil.IsKeyJustPressed(ebiten.KeyEscape) || pressed(ebiten.StandardGamepadButtonRightRight)
	activate := pressed(ebiten.StandardGamepadButtonRightBottom)
	keyActivate := inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace)
	tab := inpututil.IsKeyJustPressed(ebiten.KeyTab)
	delta := image.Point{}
	modified := ebiten.IsKeyPressed(ebiten.KeyAlt) || ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	for _, d := range directions {
		if !modified && navigationRepeat(inpututil.KeyPressDuration(d.key)) || pad >= 0 && navigationRepeat(inpututil.StandardGamepadButtonPressDuration(pad, d.button)) {
			delta = d.delta
			break
		}
	}
	step := 0
	if tab {
		step = 1
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			step = -1
		}
	}
	g.routeNavigation(delta, back, activate, keyActivate, step)
}

func (g *Game) routeNavigation(delta image.Point, back, activate, keyActivate bool, step int) {
	tab := step != 0
	if !back && !activate && !keyActivate && !tab && delta == (image.Point{}) {
		return
	}
	if g.actions.mode == "" && g.galleryInput(delta, back, activate || keyActivate, tab) {
		return
	}
	f := g.ui.GetFocusedWidget()
	// A modal owns its input, including before it has assigned initial focus.
	if g.ui.HasFocus() && (f == nil || !g.ownsFocus(f)) {
		return
	}
	field, editing := f.(*widget.TextInput)
	if editing {
		if !back && !tab {
			return
		}
		g.setNavTarget(field)
		if back {
			return
		}
	} else if f != nil {
		if keyActivate {
			return // EbitenUI already dispatched this key to its focused control.
		}
		g.setNavTarget(f)
	}
	if g.nav.resizing {
		if back || activate || keyActivate || tab {
			g.nav.resizing = false
		} else if delta.X != 0 {
			g.splitWidth += delta.X * 16
			g.clampSplitWidth()
			g.layoutDirty = true
		}
		return
	}
	if tab {
		g.stepFocus(step)
	} else if back {
		id := g.shell.Snapshot().ActiveTab().ID()
		if g.actions.mode != "" {
			g.backActionScreen()
		} else if g.inWorkspace() {
			g.workspaceBack()
		} else if g.collapsed() && g.detailOpen[id] {
			g.showList(id)
		} else {
			g.focusStrip()
		}
	} else if delta != (image.Point{}) {
		g.moveFocus(delta)
	} else if activate || keyActivate {
		g.activateFocus()
	}
}
