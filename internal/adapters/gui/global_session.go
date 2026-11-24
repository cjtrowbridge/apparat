//go:build gui

// Places the in-memory global prompt below the empty conversation area.
// Uses the same retained editor and focus route as other HUD forms.
package gui

import (
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
)

func (g *Game) globalSessionBody(tab hud.Tab) *widget.Container {
	if g.globalField != nil {
		g.globalDraft = g.globalField.GetText()
	}
	body := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewAnchorLayout()))
	form := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true, VerticalPosition: widget.AnchorLayoutPositionEnd,
			Padding: &widget.Insets{Bottom: pttMargin + 2*pttRadius + pttTravel + bodyGap},
		})))
	for _, section := range tab.Sections {
		for _, row := range section.Rows {
			g.globalField = g.projectInput(form, row.Label, g.globalDraft, func(text string) {
				g.globalDraft = text
			})
		}
	}
	body.AddChild(form)
	return body
}

func (g *Game) focusGlobalPrompt() {
	g.nav.active = true
	g.nav.strip = false
	g.nav.gallery = false
	g.ui.SetFocusedWidget(g.globalField)
	g.nav.target = focusTarget{w: g.globalField}
}

func (g *Game) openGlobalPrompt() {
	if g.actions.mode != "" {
		g.closeActionScreen()
	}
	err := g.shell.SelectTabByID(hud.TabGlobalSession)
	if err != nil {
		return
	}
	g.requestActiveTabVisible()
	g.rebuildUI(g.shell.Snapshot())
	g.focusGlobalPrompt()
	g.activeTabID.Store(string(hud.TabGlobalSession))
	if g.voice.canvas != nil && g.voice.tap.source == "floating-ptt" {
		g.voice.field = g.globalField
		g.voice.original = g.globalField.GetText()
	}
}

// Accept a released utterance from the Android dialog on the HUD thread.
func (g *Game) receiveNativePrompt() {
	v := &g.voice
	v.nativeMu.Lock()
	requested, capture := v.globalRequested, v.globalCapture
	v.globalRequested = false
	v.globalCapture = nil
	v.nativeMu.Unlock()
	if !requested {
		return
	}
	g.cancelPTT()
	g.openGlobalPrompt()
	if capture != nil {
		v.canvas = capture
		v.field = g.globalField
		v.original = v.field.GetText()
	}
}
