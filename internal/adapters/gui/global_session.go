//go:build gui

// Places the in-memory global prompt below the empty conversation area.
// Uses the same retained editor and focus route as other HUD forms.
package gui

import (
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) globalSessionBody(tab hud.Tab) *widget.Container {
	body := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewAnchorLayout()))
	form := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true, VerticalPosition: widget.AnchorLayoutPositionEnd,
			Padding: &widget.Insets{Bottom: pttMargin + 2*pttRadius + pttTravel + bodyGap},
		})))
	for _, section := range tab.Sections {
		for _, row := range section.Rows {
			surface := g.multilineInput(form, row.Label, &g.globalDraft)
			g.globalField = surface.TextInput
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
	if !g.openingVoicePrompt {
		g.editGlobalPrompt()
	}
}

func (g *Game) openGlobalPrompt() {
	if g.actions.mode != "" {
		g.closeActionScreen()
	}
	err := g.shell.SelectTabByID(hud.TabGlobalSession)
	if err != nil {
		return
	}
	if g.voice.tap.source == "floating-ptt" {
		g.voice.discardInline()
	}
	g.openingVoicePrompt = true
	defer func() {
		g.openingVoicePrompt = false
	}()
	g.requestActiveTabVisible()
	g.rebuildUI(g.shell.Snapshot())
	g.focusGlobalPrompt()
	g.activeTabID.Store(string(hud.TabGlobalSession))
	if g.voice.canvas != nil && g.voice.tap.source == "floating-ptt" {
		g.voice.targetField(g.globalField)
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
		v.targetField(g.globalField)
		v.tap = pttLock{source: "floating-ptt", locked: true, slideLocked: true, global: true}
		g.shell.StartVoiceCapture("floating-ptt")
	}
}

// The native multiline editor owns Android keyboard focus. Its result returns
// to the game thread, where keyboard and speech edits share the same draft.
func (g *Game) editGlobalPrompt() {
	prompt := g.runtimeInfo.PromptValue
	if prompt == nil || g.globalTyping != nil || g.openingVoicePrompt {
		return
	}
	if g.voice.canvas != nil {
		g.endPTT()
		g.cancelPTT()
	}
	initial := g.globalDraft
	done := make(chan *string, 1)
	g.globalTyping = done
	go func() {
		value, ok := prompt("Prompt", initial, false)
		if !ok {
			done <- nil
			return
		}
		done <- &value
	}()
}

func (g *Game) pollGlobalTyping() {
	select {
	case value := <-g.globalTyping:
		g.globalTyping = nil
		if value == nil || *value == g.globalDraft {
			return
		}
		e, _ := g.globalField.GetWidget().CustomData.(*documentSurface)
		if e != nil {
			e.anchor = 0
			e.caret = len(e.body)
			e.insert(*value)
		}
	default:
	}
}

func (g *Game) globalTypingPress() bool {
	if g.runtimeInfo.PromptValue == nil || g.globalField == nil || g.actions.mode != "" || g.shell.Snapshot().ActiveTab().ID() != hud.TabGlobalSession {
		return false
	}
	bounds := g.globalField.GetWidget().Rect
	for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
		x, y := ebiten.TouchPosition(id)
		if imagePoint(x, y).In(bounds) {
			g.editGlobalPrompt()
			return true
		}
	}
	x, y := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && imagePoint(x, y).In(bounds) {
		g.editGlobalPrompt()
		return true
	}
	return false
}
