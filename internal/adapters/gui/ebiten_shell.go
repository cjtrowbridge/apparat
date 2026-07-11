//go:build gui

package gui

import (
	"fmt"
	"image/color"
	"sync/atomic"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	backgroundColor = color.RGBA{R: 15, G: 18, B: 28, A: 255}
	panelColor      = color.RGBA{R: 25, G: 30, B: 46, A: 255}
	activeTabColor  = color.RGBA{R: 62, G: 96, B: 150, A: 255}
	mutedTextColor  = color.RGBA{R: 148, G: 160, B: 184, A: 255}
)

const (
	windowMargin      = 12
	tabTop            = 10
	tabHeight         = 54
	tabGap            = 6
	tabTextInsetX     = 18
	tabTextInsetY     = 18
	debugGlyphWidth   = 6
	tabDragThreshold  = 18
	bodyGap           = 8
	diagnosticsHeight = 58
	collapsedWidth    = 760
	defaultSplitWidth = 260
	minSplitWidth     = 170
	maxSplitWidth     = 520
)

type Game struct {
	shell            hud.Shell
	ui               *ebitenui.UI
	theme            *widget.Theme
	width            int
	height           int
	layoutDirty      bool
	rightCtrlHeld    bool
	pageDownWasHit   bool
	pageUpWasHit     bool
	l1WasPressed     bool
	r1WasPressed     bool
	r2Held           bool
	debugOverlayOpen bool
	debugOverlayX    int
	debugOverlayY    int
	debugDrag        bool
	debugDragDX      int
	debugDragDY      int
	debugTouchActive bool
	debugTouchID     ebiten.TouchID
	tabScroll        *widget.ScrollContainer
	tabButtonCount   int
	tabStripDragging bool
	tabStripLastX    int
	tabTouchActive   bool
	tabTouchID       ebiten.TouchID
	selectedSections map[hud.TabID]int
	detailOpen       map[hud.TabID]bool
	splitWidth       int
	splitDragging    bool
	runtimeInfo      RuntimeInfo
	activeTabID      atomic.Value
	updateStatus     atomic.Value
	updateButton     *widget.Button
	onCheckForUpdate func() bool
}

func (game *Game) SetOnCheckForUpdate(f func() bool) {
	game.onCheckForUpdate = f
}

func (game *Game) SetUpdateStatus(status string) {
	game.updateStatus.Store(status)
}

func (game *Game) UpdateStatus() string {
	if status, ok := game.updateStatus.Load().(string); ok {
		return status
	}
	return ""
}

func NewGame() *Game {
	return NewGameWithRuntimeInfo(defaultRuntimeInfo())
}

func NewGameWithRuntimeInfo(info RuntimeInfo) *Game {
	theme := createUITheme()
	game := &Game{
		shell:            hud.NewShell(),
		theme:            theme,
		width:            1280,
		height:           800,
		debugOverlayX:    24,
		debugOverlayY:    96,
		selectedSections: map[hud.TabID]int{},
		detailOpen:       map[hud.TabID]bool{},
		splitWidth:       defaultSplitWidth,
		runtimeInfo:      info,
	}
	game.rebuildUI(game.shell.Snapshot())
	game.activeTabID.Store(string(game.shell.Snapshot().ActiveTab().ID()))
	return game
}

func (game *Game) Update() error {
	if game.layoutDirty {
		game.layoutDirty = false
		game.rebuildUI(game.shell.Snapshot())
	}
	game.ui.Update()
	game.applyUpdateStatus()
	startIndex := game.shell.Snapshot().ActiveIndex
	game.updatePointerState()
	ctrlPressed := ebiten.IsKeyPressed(ebiten.KeyControl)
	pageDownPressed := ebiten.IsKeyPressed(ebiten.KeyPageDown)
	pageUpPressed := ebiten.IsKeyPressed(ebiten.KeyPageUp)
	if ctrlPressed && pageDownPressed && !game.pageDownWasHit {
		_ = game.shell.ApplyAction(hud.ActionNextTab)
	}
	if ctrlPressed && pageUpPressed && !game.pageUpWasHit {
		_ = game.shell.ApplyAction(hud.ActionPreviousTab)
	}
	game.pageDownWasHit = pageDownPressed
	game.pageUpWasHit = pageUpPressed

	for index, key := range []ebiten.Key{ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4, ebiten.Key5, ebiten.Key6, ebiten.Key7} {
		if ebiten.IsKeyPressed(ebiten.KeyAlt) && ebiten.IsKeyPressed(key) {
			_ = game.shell.SelectTab(index)
		}
	}
	rightCtrl := ebiten.IsKeyPressed(ebiten.KeyControlRight)
	if rightCtrl && !game.rightCtrlHeld {
		game.shell.StartVoiceCapture("right-ctrl")
	}
	if !rightCtrl && game.rightCtrlHeld {
		game.shell.ReleaseVoiceCapture()
	}
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		game.shell.CancelVoiceCapture()
	}
	game.rightCtrlHeld = rightCtrl
	game.updateGamepad()
	if activeIndex := game.shell.Snapshot().ActiveIndex; activeIndex != startIndex {
		game.rebuildUI(game.shell.Snapshot())
	}
	game.ensureActiveTabVisible()
	game.activeTabID.Store(string(game.shell.Snapshot().ActiveTab().ID()))
	return nil
}

func (game *Game) applyUpdateStatus() {
	if game.updateButton == nil {
		return
	}
	status := game.UpdateStatus()
	if status == "" {
		return
	}
	text := game.updateButton.Text()
	if text != nil && text.Label == status {
		return
	}
	game.updateButton.SetText(status)
}

func (game *Game) ActiveTabID() string {
	if id, ok := game.activeTabID.Load().(string); ok {
		return id
	}
	return ""
}

func (game *Game) updateGamepad() {
	for _, id := range ebiten.AppendGamepadIDs(nil) {
		if !ebiten.IsStandardGamepadLayoutAvailable(id) {
			continue
		}
		l1 := ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonFrontTopLeft)
		r1 := ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonFrontTopRight)
		r2 := ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonFrontBottomRight)
		if r1 && !game.r1WasPressed {
			_ = game.shell.ApplyAction(hud.ActionNextTab)
		}
		if l1 && !game.l1WasPressed {
			_ = game.shell.ApplyAction(hud.ActionPreviousTab)
		}
		if r2 && !game.r2Held {
			game.shell.StartVoiceCapture("gamepad-r2")
		}
		if !r2 && game.r2Held {
			game.shell.ReleaseVoiceCapture()
		}
		game.l1WasPressed = l1
		game.r1WasPressed = r1
		game.r2Held = r2
		break // Use only the first standard-layout gamepad.
	}
}

func (game *Game) Draw(screen *ebiten.Image) {
	screen.Fill(backgroundColor)
	if game.ui != nil {
		game.ui.Draw(screen)
	}
	if game.debugOverlayOpen {
		game.drawDebugOverlay(screen)
	}
	snapshot := game.shell.Snapshot()
	drawDiagnostics(screen, snapshot, game.height)
}

func drawDiagnostics(screen *ebiten.Image, snapshot hud.Snapshot, height int) {
	y := height - diagnosticsHeight + 6
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("voice=%s route=%s input=%s", snapshot.VoiceState, snapshot.Diagnostics.ActiveRoute, snapshot.Diagnostics.Input), windowMargin, y)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("config tabPlacement=%s theme=%s scale=%.1f", snapshot.Config.TabView.Placement, snapshot.Config.Display.Theme, snapshot.Config.Display.Scale), windowMargin, y+18)
	ebitenutil.DebugPrintAt(screen, "L1/R1 or Ctrl+PageUp/PageDown switch tabs • R2/right Ctrl push-to-talk", windowMargin, y+36)
	_ = mutedTextColor
}

func (game *Game) Layout(outsideWidth int, outsideHeight int) (int, int) {
	if outsideWidth > 0 {
		game.layoutDirty = game.layoutDirty || outsideWidth != game.width
		game.width = outsideWidth
	}
	if outsideHeight > 0 {
		game.layoutDirty = game.layoutDirty || outsideHeight != game.height
		game.height = outsideHeight
	}
	return game.width, game.height
}
