//go:build gui

// Updates and draws the Ebitengine HUD shell.
// Dependencies: github.com/cjtrowbridge/apparat/internal/hud, github.com/ebitenui/ebitenui, github.com/ebitenui/ebitenui/widget.
package gui

import (
	"context"
	"fmt"
	"image/color"
	"sync/atomic"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	ebitentext "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
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
	pttRadius         = 36
	pttMargin         = 22
	collapsedWidth    = 760
	defaultSplitWidth = 260
	minSplitWidth     = 170
	maxSplitWidth     = 520
)

type Game struct {
	prefs                          preferenceState
	projects                       projectForm
	window                         windowTracker
	arcade                         arcadeView
	shell                          hud.Shell
	runContext                     context.Context
	gesture                        pointerGesture
	cancelInputRequested           atomic.Bool
	serviceUpdates                 <-chan ServiceProjection
	serviceProjection              ServiceProjection
	cancelQueries                  context.CancelFunc
	ui                             *ebitenui.UI
	theme                          *widget.Theme
	selectorDescriptionFace        *ebitentext.Face
	pageTitleFace                  *ebitentext.Face
	width                          int
	height                         int
	layoutDirty                    bool
	rightCtrlHeld                  bool
	pageDownWasHit                 bool
	pageUpWasHit                   bool
	l1WasPressed                   bool
	r1WasPressed                   bool
	r2Held                         bool
	debugOverlayOpen               bool
	debugOverlayX                  int
	debugOverlayY                  int
	debugDragDX                    int
	debugDragDY                    int
	tabScroll                      *widget.ScrollContainer
	tabButtonCount                 int
	tabStripContentWidth           int
	tabButtonLefts                 []int
	tabButtonRights                []int
	tabStripScrollLeft             float64
	tabButtons                     []*widget.Button
	tabStripDragDistance           int
	tabStripDragMoved              bool
	tabStripLastX                  int
	tabRadioGroup                  *widget.RadioGroup
	tabDragCancelUpdates           int
	verticalScrolls                []*widget.ScrollContainer
	bodyScroll                     *widget.ScrollContainer
	bodyScrollLastY                int
	bodyScrollDragDistance         int
	bodyScrollDragMoved            bool
	bodyScrollCancelUpdates        int
	bodyTabScrollLeft              float64
	bodyTabScrollLocked            bool
	syncingTabButtonStates         bool
	activeTabScrollPending         bool
	selectedSections               map[hud.TabID]int
	detailOpen                     map[hud.TabID]bool
	splitWidth                     int
	divider                        *widget.Widget
	splitGrabOffset                int
	researchContributionSort       hud.ContributionSort
	researchContributionDescending bool
	pttHeld                        bool
	runtimeInfo                    RuntimeInfo
	activeTabID                    atomic.Value
	updateStatus                   atomic.Value
	updateButton                   *widget.Button
	updateDetail                   *widget.Text
	onCheckForUpdate               func() bool
}

func (game *Game) SetOnCheckForUpdate(f func() bool) {
	game.onCheckForUpdate = f
}

func (game *Game) SetUpdateStatus(code UpdateStatusCode) {
	game.updateStatus.Store(updateStatusFor(code))
}

func (game *Game) UpdateStatus() string {
	{
		status, ok := game.updateStatus.Load().(updateStatus)
		if ok {
			return status.Label
		}
	}
	return ""
}

func (game *Game) UpdateStatusDetail() string {
	{
		status, ok := game.updateStatus.Load().(updateStatus)
		if ok {
			return status.Detail
		}
	}
	return ""
}

func NewGame() *Game {
	return NewGameWithRuntimeInfo(defaultRuntimeInfo())
}

func NewGameWithRuntimeInfo(info RuntimeInfo) *Game {
	theme := createUITheme()
	game := &Game{
		shell:                          hud.NewShell(),
		theme:                          theme,
		selectorDescriptionFace:        createTextFace(11),
		pageTitleFace:                  createTextFace(28),
		width:                          1280,
		height:                         800,
		debugOverlayX:                  24,
		debugOverlayY:                  96,
		selectedSections:               map[hud.TabID]int{},
		detailOpen:                     map[hud.TabID]bool{},
		splitWidth:                     defaultSplitWidth,
		researchContributionSort:       hud.ContributionSortGFlops,
		researchContributionDescending: true,
		runtimeInfo:                    info,
	}
	if info.Queries != nil {
		ctx, cancel := context.WithCancel(context.Background())
		game.cancelQueries = cancel
		game.serviceProjection.State = "loading"
		game.serviceUpdates = WatchServices(ctx, info.Queries)
	}
	game.loadPreferences()
	game.initProjects()
	game.rebuildUI(game.shell.Snapshot())
	game.activeTabID.Store(string(game.shell.Snapshot().ActiveTab().ID()))
	return game
}

func (game *Game) Update() error {
	if game.runContext != nil && game.runContext.Err() != nil {
		return ebiten.Termination
	}
	if game.cancelInputRequested.Swap(false) {
		game.cancelPointer()
	}
	game.receivePreferences()
	game.pollProjects()
	game.trackWindow()
	game.receiveServices()
	if game.layoutDirty && game.gesture.Owner == gestureNone {
		game.layoutDirty = false
		game.rebuildUI(game.shell.Snapshot())
	}
	if game.arcadeUpdate() {
		return nil
	}
	game.updatePointerState()
	game.ui.Update()
	game.enforceBodyTabScrollLock()
	game.advanceTabDragCancellation()
	game.advanceBodyScrollCancellation()
	game.applyUpdateStatus()
	startIndex := game.shell.Snapshot().ActiveIndex
	ctrlPressed := ebiten.IsKeyPressed(ebiten.KeyControl)
	pageDownPressed := ebiten.IsKeyPressed(ebiten.KeyPageDown)
	pageUpPressed := ebiten.IsKeyPressed(ebiten.KeyPageUp)
	if ctrlPressed && pageDownPressed && !game.pageDownWasHit {
		game.applyTabAction(hud.ActionNextTab)
	}
	if ctrlPressed && pageUpPressed && !game.pageUpWasHit {
		game.applyTabAction(hud.ActionPreviousTab)
	}
	game.pageDownWasHit = pageDownPressed
	game.pageUpWasHit = pageUpPressed

	for index, key := range []ebiten.Key{ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4, ebiten.Key5, ebiten.Key6} {
		if ebiten.IsKeyPressed(ebiten.KeyAlt) && ebiten.IsKeyPressed(key) {
			game.selectTabWithReveal(index)
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
		game.setPTTHeld(false)
	}
	game.rightCtrlHeld = rightCtrl
	game.updateGamepad()
	{
		activeIndex := game.shell.Snapshot().ActiveIndex
		if activeIndex != startIndex {
			game.rebuildUI(game.shell.Snapshot())
		}
	}
	game.ensureActiveTabVisible()
	game.activeTabID.Store(string(game.shell.Snapshot().ActiveTab().ID()))
	return nil
}

func (game *Game) applyUpdateStatus() {
	status := game.UpdateStatus()
	if status == "" {
		return
	}
	if game.updateButton != nil {
		text := game.updateButton.Text()
		if text == nil || text.Label != status {
			game.updateButton.SetText(status)
		}
	}
	if game.updateDetail != nil {
		game.updateDetail.Label = game.UpdateStatusDetail()
	}
}

func (game *Game) ActiveTabID() string {
	{
		id, ok := game.activeTabID.Load().(string)
		if ok {
			return id
		}
	}
	return ""
}

func (game *Game) updateGamepad() {
	for _, id := range ebiten.AppendGamepadIDs(nil) {
		if !ebiten.IsStandardGamepadLayoutAvailable(id) {
			continue
		}
		game.projectPad(id)
		l1 := ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonFrontTopLeft)
		r1 := ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonFrontTopRight)
		r2 := ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonFrontBottomRight)
		if r1 && !game.r1WasPressed {
			game.applyTabAction(hud.ActionNextTab)
		}
		if l1 && !game.l1WasPressed {
			game.applyTabAction(hud.ActionPreviousTab)
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

func (game *Game) applyTabAction(action hud.Action) {
	active := game.shell.Snapshot().ActiveIndex
	{
		err := game.shell.ApplyAction(action)
		if err == nil && game.shell.Snapshot().ActiveIndex != active {
			game.requestActiveTabVisible()
		}
	}
}

func (game *Game) selectTabWithReveal(index int) {
	active := game.shell.Snapshot().ActiveIndex
	{
		err := game.shell.SelectTab(index)
		if err == nil && game.shell.Snapshot().ActiveIndex != active {
			game.requestActiveTabVisible()
		}
	}
}

func (game *Game) Draw(screen *ebiten.Image) {
	screen.Fill(backgroundColor)
	if game.ui != nil {
		game.ui.Draw(screen)
	}
	if game.arcadeDraw(screen) {
		return
	}
	game.drawPTTButton(screen)
	if game.debugOverlayOpen {
		game.drawDebugOverlay(screen)
	}
	snapshot := game.shell.Snapshot()
	drawDiagnostics(screen, snapshot, game.height)
}

func drawDiagnostics(screen *ebiten.Image, snapshot hud.Snapshot, height int) {
	y := height - diagnosticsHeight + 6
	if screen.Bounds().Dx() < collapsedWidth {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("voice=%s input=%s", snapshot.VoiceState, snapshot.Diagnostics.Input), windowMargin, y)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("tab=%s theme=%s", snapshot.ActiveTab().ID(), snapshot.Config.Display.Theme), windowMargin, y+18)
		ebitenutil.DebugPrintAt(screen, "Swipe tabs | Hold PTT | Ctrl+PgUp/PgDn", windowMargin, y+36)
		return
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("voice=%s route=%s input=%s", snapshot.VoiceState, snapshot.Diagnostics.ActiveRoute, snapshot.Diagnostics.Input), windowMargin, y)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("config tabPlacement=%s theme=%s scale=%.1f", snapshot.Config.TabView.Placement, snapshot.Config.Display.Theme, snapshot.Config.Display.Scale), windowMargin, y+18)
	ebitenutil.DebugPrintAt(screen, "L1/R1 or Ctrl+PageUp/PageDown switch tabs • R2/right Ctrl push-to-talk", windowMargin, y+36)
	_ = mutedTextColor
}

func (game *Game) drawPTTButton(screen *ebiten.Image) {
	x, y := game.pttCenter()
	fill := color.RGBA{R: 38, G: 90, B: 156, A: 230}
	if game.pttHeld {
		fill = color.RGBA{R: 30, G: 150, B: 110, A: 240}
	}
	vector.DrawFilledCircle(screen, float32(x), float32(y), float32(pttRadius), fill, true)
	vector.StrokeCircle(screen, float32(x), float32(y), float32(pttRadius), 2, color.RGBA{R: 220, G: 232, B: 255, A: 255}, true)
	ebitenutil.DebugPrintAt(screen, "PTT", x-10, y-4)
}

func (game *Game) pointInPTT(x int, y int) bool {
	cx, cy := game.pttCenter()
	dx, dy := x-cx, y-cy
	return dx*dx+dy*dy <= pttRadius*pttRadius
}

func (game *Game) setPTTHeld(held bool) {
	if held == game.pttHeld {
		return
	}
	game.pttHeld = held
	if held {
		game.shell.StartVoiceCapture("floating-ptt")
		return
	}
	game.shell.ReleaseVoiceCapture()
}

func (game *Game) Layout(outsideWidth int, outsideHeight int) (int, int) {
	outsideWidth = outsideWidth * 100 / max(1, game.prefs.Scale)
	outsideHeight = outsideHeight * 100 / max(1, game.prefs.Scale)
	changed := false
	if outsideWidth > 0 {
		changed = changed || outsideWidth != game.width
		game.layoutDirty = game.layoutDirty || changed
		game.width = outsideWidth
	}
	if outsideHeight > 0 {
		changed = changed || outsideHeight != game.height
		game.layoutDirty = game.layoutDirty || changed
		game.height = outsideHeight
	}
	if changed {
		game.cancelPointer()
		game.requestActiveTabVisible()
	}
	return game.width, game.height
}
