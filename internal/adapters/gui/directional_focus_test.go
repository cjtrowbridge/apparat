//go:build gui

// Exercises directional traversal on the rendered HUD and embedded gallery.
// These tests cover ownership, scroll reveal, restoration and tab orientation.
package gui

import (
	"image"
	"testing"

	"ebe-boilerplate/surface"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

func renderNavigation(g *Game) {
	dst := ebiten.NewImage(g.width, g.height)
	defer dst.Deallocate()
	g.ui.Container.SetLocation(dst.Bounds())
	g.ui.Draw(dst)
	g.drawFocus(dst)
}

func TestDirectionalStripAndPageEntry(t *testing.T) {
	for _, placement := range []string{"left_rail", "right_rail", "top", "bottom"} {
		g := NewGame()
		g.prefs.Placement = placement
		for index, tab := range g.shell.Snapshot().Tabs {
			g.selectTab(index, nil)
			renderNavigation(g)
			g.focusStrip()
			tangent := image.Pt(1, 0)
			if g.sideRail() != "" {
				tangent = image.Pt(0, 1)
			}
			g.moveFocus(tangent)
			count := len(g.tabButtons)
			if g.shell.Snapshot().ActiveIndex != (index+1)%count {
				t.Fatalf("%s: failed to switch from %s", placement, tab.ID())
			}
			g.selectTab(index, nil)
			renderNavigation(g)
			g.moveFocus(g.inward())
			if g.nav.strip {
				t.Fatalf("%s: could not enter %s", placement, tab.ID())
			}
			if tab.ID() == "stmpck" && !g.nav.gallery {
				t.Fatal("gallery did not acquire focus")
			}
		}
	}
}

func TestSettingsDirectionalRevealAndRestore(t *testing.T) {
	g := horizontalGame()
	selectTabThroughStrip(t, g, hud.TabSettings)
	g.Layout(640, 400)
	g.rebuildUI(g.shell.Snapshot())
	renderNavigation(g)
	g.focusStrip()
	g.moveFocus(g.inward())
	targets := g.bodyTargets()
	for _, target := range targets[1:] {
		g.moveFocus(image.Pt(0, 1))
		if g.nav.target.w != target.w {
			t.Fatalf("down missed %q", focusLabel(target.w))
		}
		renderNavigation(g)
	}
	if g.verticalScrolls[0].ScrollTop <= 0 {
		t.Fatal("last control was not revealed")
	}
	before := focusLabel(g.nav.target.w)
	g.rebuildUI(g.shell.Snapshot())
	renderNavigation(g)
	if g.nav.target.w == nil || focusLabel(g.nav.target.w) != before {
		t.Fatal("rebuild lost focus")
	}
}

func TestDirectionalEditingAndDisabledControl(t *testing.T) {
	g := horizontalGame()
	c := g.chatComposer()
	g.nav.body = c
	field := c.Children()[0].(*widget.TextInput)
	field.SetText("draft")
	g.enterBody()
	if field.IsFocused() {
		t.Fatal("visiting a field entered editing")
	}
	g.routeNavigation(image.Point{}, false, true, false, 0)
	g.routeNavigation(image.Pt(1, 0), false, false, false, 0)
	if g.ui.GetFocusedWidget() != field {
		t.Fatal("activation did not enter editing")
	}
	g.routeNavigation(image.Point{}, true, false, false, 0)
	if field.IsFocused() || field.GetText() != "draft" {
		t.Fatal("leaving editing changed draft or retained caret ownership")
	}
	c.Children()[1].GetWidget().Disabled = true
	if len(g.bodyTargets()) != 1 {
		t.Fatal("disabled control is reachable")
	}
}

func TestGalleryDirectionalOwnership(t *testing.T) {
	g := horizontalGame()
	selectTabThroughStrip(t, g, "stmpck")
	s := surface.New(t.TempDir())
	g.arcade.session.Store(s)
	g.arcade.canvas = ebiten.NewImage(640, 480)
	defer g.arcade.canvas.Deallocate()
	s.Layout(640, 480)
	g.focusStrip()
	right := image.Pt(1, 0)
	down := image.Pt(0, 1)
	if g.galleryInput(right, false, false, false) {
		t.Fatal("gallery consumed a strip direction")
	}
	g.moveFocus(down)
	for range 4 {
		g.galleryInput(right, false, false, false)
	}
	g.galleryInput(image.Point{}, false, true, false)
	if !s.Settings {
		t.Fatal("toolbar gear did not open settings")
	}
	g.galleryInput(down, false, false, false)
	if g.nav.setting != 1 || !s.Settings {
		t.Fatal("settings did not trap focus")
	}
	g.galleryInput(image.Point{}, true, false, false)
	g.galleryInput(down, false, false, false)
	g.galleryInput(right, false, false, false)
	g.galleryInput(down, false, false, false)
	if s.Focus != 4 || s.Selected >= 0 {
		t.Fatal("directional grid movement launched or missed a game")
	}
	g.galleryInput(image.Point{}, false, true, false)
	if s.Selected != s.Order()[4] {
		t.Fatal("activation did not open focused game")
	}
}

func TestArrowAndPadDirectionBindings(t *testing.T) {
	keys := []ebiten.Key{ebiten.KeyArrowUp, ebiten.KeyArrowDown, ebiten.KeyArrowLeft, ebiten.KeyArrowRight}
	buttons := []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonLeftTop, ebiten.StandardGamepadButtonLeftBottom, ebiten.StandardGamepadButtonLeftLeft, ebiten.StandardGamepadButtonLeftRight}
	want := []image.Point{image.Pt(0, -1), image.Pt(0, 1), image.Pt(-1, 0), image.Pt(1, 0)}
	for i, key := range keys {
		found := false
		for _, binding := range directions {
			if binding.key == key && binding.button == buttons[i] && binding.delta == want[i] {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing keyboard/controller pair for %v", want[i])
		}
	}
}

func TestDirectionalPaneDivider(t *testing.T) {
	g := horizontalGame()
	renderNavigation(g)
	for i, target := range g.bodyTargets() {
		if target.w.GetWidget() != g.divider {
			continue
		}
		g.focusBody(g.bodyTargets(), i)
		g.routeNavigation(image.Point{}, false, false, true, 0)
		before := g.splitWidth
		g.routeNavigation(image.Pt(1, 0), false, false, false, 0)
		if !g.nav.resizing || g.splitWidth <= before {
			t.Fatal("focused divider did not resize")
		}
		g.routeNavigation(image.Point{}, true, false, false, 0)
		if g.nav.resizing {
			t.Fatal("Back did not finish resizing")
		}
		return
	}
	t.Fatal("missing divider")
}

func TestDirectionalActivationOnce(t *testing.T) {
	for _, keyboard := range []bool{true, false} {
		g := horizontalGame()
		selectTabThroughStrip(t, g, hud.TabSettings)
		renderNavigation(g)
		targets := g.bodyTargets()
		for i, target := range targets {
			check, ok := target.w.(*widget.Checkbox)
			if !ok {
				continue
			}
			g.focusBody(targets, i)
			before := check.State()
			g.routeNavigation(image.Point{}, false, !keyboard, keyboard, 0)
			if check.State() == before {
				t.Fatal("checkbox failed to toggle")
			}
			event.ExecuteDeferred()
			break
		}
		g.selectTab(0, nil)
		renderNavigation(g)
		g.focusStrip()
		g.moveFocus(g.inward())
		if len(g.selectedSections) > 0 {
			t.Fatal("entering page activated selector")
		}
		g.routeNavigation(image.Point{}, false, !keyboard, keyboard, 0)
		event.ExecuteDeferred()
		_, selected := g.selectedSections[g.nav.owner]
		if !selected {
			t.Fatal("activation did not select the focused section")
		}
	}
}

func TestDirectionalChatEditingSurvivesRefresh(t *testing.T) {
	g := horizontalGame()
	g.selectTab(0, nil)
	g.selectSection(hud.TabComrades, 1)
	renderNavigation(g)
	for i, target := range g.bodyTargets() {
		field, ok := target.w.(*widget.TextInput)
		if !ok {
			continue
		}
		g.focusBody(g.bodyTargets(), i)
		g.activateFocus()
		field.SetText("unsent draft")
		field.SelectAll()
		g.rebuildUI(g.shell.Snapshot())
		if g.ui.GetFocusedWidget() != field || field.SelectedText() != "unsent draft" {
			t.Fatal("refresh lost the active field, selection or draft")
		}
		g.selectSection(hud.TabComrades, 3)
		if g.ui.GetFocusedWidget() == field {
			t.Fatal("another conversation inherited the editor")
		}
		return
	}
	t.Fatal("missing composer")
}

func TestCompactGalleryToolbarAndFirstTile(t *testing.T) {
	s := surface.New(t.TempDir())
	for _, width := range []int{320, 360, 800, 1280} {
		w, h := s.Layout(width, 800)
		s.SetTab(0)
		var previous image.Rectangle
		for i := range galleryToolbarItems {
			r := s.ToolbarRect(i)
			if r.Min.Y != 0 || r.Empty() || r.Max.X > w || r.Min.X < previous.Max.X || r.Dy() < 44 {
				t.Fatalf("width %d toolbar control %d bounds %v after %v", w, i, r, previous)
			}
			previous = r
		}
		gear := s.ToolbarRect(4)
		s.ActivateMenu(float64(gear.Min.X+1), 1)
		if !s.Settings {
			t.Fatal("top-row gear did not open settings")
		}
		s.Back()
		x, y, _, _ := s.TileRect(0, w)
		if y != 48 {
			t.Fatalf("first tile starts at %v", y)
		}
		s.ActivateMenu(x+1, y+1)
		if s.Selected < 0 {
			t.Fatal("first tile did not open from its new top edge")
		}
		s.Back()
		s.Move(len(s.Order()), h)
		x, y, _, tileHeight := s.TileRect(s.Focus, w)
		if y < 48 || y+tileHeight > float64(h) || s.TileAt(x+1, y+1, w) != s.Focus {
			t.Fatalf("last tile is clipped or misses its target at width %d", w)
		}
		s.Move(-len(s.Order()), h)
	}
}
