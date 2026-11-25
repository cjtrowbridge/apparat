//go:build gui

// Routes voice entry and keyboard focus to the current planning conversation.
// Uses the same retained editor and focus route as other HUD forms.
package gui

import (
	"github.com/cjtrowbridge/apparat/internal/hud"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) focusGlobalPrompt() {
	if g.globalField == nil {
		return
	}
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
	g.planning.page = ""
	err := g.shell.SelectTabByID(hud.TabGlobalSession)
	if err != nil {
		return
	}
	if g.voice.tap.source == "floating-ptt" {
		g.voice.discardInline()
	}
	g.globalVoiceFocus = true
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
	g.enterPlanningVoice()
	if capture != nil {
		v.canvas = capture
		v.targetField(g.globalField)
		v.tap = pttLock{source: "floating-ptt", locked: true, slideLocked: true, global: true}
		g.shell.StartVoiceCapture("floating-ptt")
	}
}

// Arm the canvas IME only for manual prompt entry. Voice focus uses the
// same field but keeps the keyboard dismissed during and after capture.
func (g *Game) editGlobalPrompt() {
	if g.globalField == nil {
		return
	}
	if g.voice.canvas != nil {
		g.endPTT()
		g.cancelPTT()
	}
	g.globalVoiceFocus = false
	g.ui.SetFocusedWidget(g.globalField)
}

func (g *Game) globalTypingPress() {
	_, terminal := g.currentChatTerminal()
	if terminal != nil && terminal.permissionPage {
		return
	}
	if g.globalField == nil || g.actions.mode != "" || g.shell.Snapshot().ActiveTab().ID() != hud.TabGlobalSession {
		return
	}
	bounds := g.globalField.GetWidget().Rect
	for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
		x, y := ebiten.TouchPosition(id)
		// Releasing text focus resizes the Android canvas before a touch-up can
		// reach this same button. Resolve the tap in the pre-resize geometry.
		button := g.planning.modelButton
		if g.runtimeInfo.CanvasFitsKeyboard && g.globalField.IsFocused() && button != nil && imagePoint(x, y).In(button.GetWidget().Rect) {
			g.openModels()
			return
		}
		if imagePoint(x, y).In(bounds) {
			g.editGlobalPrompt()
			return
		}
	}
	x, y := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && imagePoint(x, y).In(bounds) {
		g.editGlobalPrompt()
	}
}

func (game *Game) leaveGlobalPrompt() {
	if game.globalField != nil && (game.actions.mode != "" || game.shell.Snapshot().ActiveTab().ID() != hud.TabGlobalSession) {
		game.globalField.GetWidget().CustomData.(*documentSurface).ime.Confirm()
		game.globalField = nil
		game.globalVoiceFocus = false
	}
}
