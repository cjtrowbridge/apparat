//go:build gui

// Applies confirmed preferences on the UI thread and saves edits off-thread.
// The initial snapshot is loaded before constructing the visible widget tree.
package gui

import (
	"context"
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"image/color"
	"slices"
)

type preferenceStore interface {
	Preferences(context.Context) (domain.Preferences, error)
	SavePreferences(context.Context, domain.Preferences) error
}
type preferenceState struct {
	reveal       string
	revealWidget *widget.Widget
	domain.Preferences
	pane    *widget.ScrollContainer
	store   preferenceStore
	pending chan error
	next    domain.Preferences
	message string
}

func (g *Game) loadPreferences() {
	p := domain.DefaultPreferences()
	store, ok := g.runtimeInfo.Queries.(preferenceStore)
	if ok {
		g.prefs.store = store
		loaded, err := store.Preferences(context.Background())
		if err == nil {
			p = loaded
		} else {
			g.prefs.message = "Could not read saved settings"
		}
	}
	g.applyPreferences(p)
	for i, tab := range g.shell.Snapshot().Tabs {
		if string(tab.ID()) == p.Landing {
			g.selectTabWithReveal(i)
			break
		}
	}
}
func (g *Game) applyPreferences(p domain.Preferences) {
	g.prefs.Preferences = p
	g.debugOverlayOpen = p.Overlay
	g.theme = createUITheme()
	face := createTextFace(float64(p.Font))
	g.theme.ButtonTheme.TextFace = face
	g.theme.LabelTheme.Face = face
	g.pageTitleFace = createTextFace(float64(p.Font) * 2)
	g.selectorDescriptionFace = createTextFace(float64(p.Font) * 0.8)
	theme := hud.ThemeDark
	if p.Theme == "High contrast" {
		theme = hud.ThemeHighContrast
		g.theme.PanelTheme.BackgroundImage = uiimage.NewNineSliceColor(color.Black)
		g.theme.ButtonTheme.TextColor.Idle = color.White
		g.theme.LabelTheme.Color.Idle = color.White
	}
	g.shell.SetTabPlacement(hud.TabPlacement(p.Placement))
	g.shell.ApplyDisplay(theme, float64(p.Scale)/100, p.Font)
}
func (g *Game) savePreferences(p domain.Preferences) {
	if g.prefs.pending != nil {
		return
	}
	if g.prefs.store == nil {
		g.applyPreferences(p)
		g.layoutDirty = true
		return
	}
	result := make(chan error, 1)
	g.prefs.pending = result
	g.prefs.next = p
	store := g.prefs.store
	go func() {
		result <- store.SavePreferences(context.Background(), p)
	}()
}
func (g *Game) receivePreferences() {
	select {
	case err := <-g.prefs.pending:
		g.prefs.pending = nil
		if err != nil {
			g.prefs.message = "Save failed. Select the setting to retry."
		} else {
			g.prefs.message = ""
			g.applyPreferences(g.prefs.next)
		}
		g.layoutDirty = true
	default:
	}
}
func cycle[T comparable](current T, values []T) T {
	return values[(slices.Index(values, current)+1)%len(values)]
}
func (g *Game) preferenceControls(c *widget.Container, section string) {
	p := g.prefs.Preferences
	add := func(label string, change func(*domain.Preferences)) {
		c.AddChild(widget.NewButton(widget.ButtonOpts.Text(label, g.theme.ButtonTheme.TextFace, g.theme.ButtonTheme.TextColor),
			widget.ButtonOpts.Image(g.theme.ButtonTheme.Image), widget.ButtonOpts.TextPadding(g.theme.ButtonTheme.TextPadding),
			widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(0, 44)),
			widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
				if g.bodySelectionSuppressed() {
					return
				}
				next := g.prefs.Preferences
				change(&next)
				g.savePreferences(next)
			})))
	}
	if section == "hud-configuration" {
		add("Theme: "+p.Theme, func(v *domain.Preferences) {
			v.Theme = cycle(v.Theme, []string{"Dark", "High contrast"})
		})
		add(fmt.Sprintf("UI scale: %d%%", p.Scale), func(v *domain.Preferences) {
			v.Scale = cycle(v.Scale, []int{75, 100, 125, 150})
		})
		add(fmt.Sprintf("Font size: %d", p.Font), func(v *domain.Preferences) {
			v.Font = cycle(v.Font, []int{14, 16, 18, 20})
		})
		add("Reset display preferences", func(v *domain.Preferences) {
			defaults := domain.DefaultPreferences()
			v.Theme = defaults.Theme
			v.Font = defaults.Font
			v.Scale = defaults.Scale
		})
	} else {
		add("Tab placement: "+placementLabel(p.Placement), func(v *domain.Preferences) {
			v.Placement = cycle(v.Placement, []string{"top", "bottom", "left_rail", "right_rail", "responsive"})
		})
		add("Startup tab: "+p.Landing, func(v *domain.Preferences) {
			var ids []string
			for _, tab := range g.shell.Snapshot().Tabs {
				ids = append(ids, string(tab.ID()))
			}
			v.Landing = cycle(v.Landing, ids)
		})
		if g.window.store != nil {
			add("Reset window placement", func(v *domain.Preferences) {
				g.resetWindow()
			})
		}
	}
	if g.prefs.message != "" {
		c.AddChild(g.hudText(g.prefs.message))
	}
}

func checkedState(checked bool) widget.WidgetState {
	if checked {
		return widget.WidgetChecked
	}
	return widget.WidgetUnchecked
}

func (g *Game) revealSetting() {
	p := &g.prefs
	if p.reveal == "" || p.pane == nil || p.revealWidget == nil {
		return
	}
	rect := p.revealWidget.Rect
	if rect.Empty() {
		return
	}
	moveBodyScroll(p.pane, float64(p.pane.ViewRect().Min.Y-rect.Min.Y))
	p.reveal = ""
	p.revealWidget = nil
}
