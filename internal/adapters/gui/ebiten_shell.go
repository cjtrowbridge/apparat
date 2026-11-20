//go:build gui

// Updates and draws the Ebitengine HUD shell.
// Dependencies: github.com/cjtrowbridge/apparat/internal/hud, github.com/ebitenui/ebitenui, github.com/ebitenui/ebitenui/widget.
package gui

import (
	"context"
	"image/color"
	"sync/atomic"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	ebitentext "github.com/hajimehoshi/ebiten/v2/text/v2"
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
	voice                          voiceState
	prefs                          preferenceState
	projects                       projectForm
	window                         windowTracker
	arcade                         arcadeView
	shell                          hud.Shell
	runContext                     context.Context
	gesture                        pointerGesture
	workspaceBackEnabled           atomic.Bool
	workspaceBackRequested         atomic.Bool
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
	discovery                      updateDiscovery
	desktopUpdates                 *desktopUpdateState
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
	game.initSpeech()
	game.rebuildUI(game.shell.Snapshot())
	game.activeTabID.Store(string(game.shell.Snapshot().ActiveTab().ID()))
	return game
}

func (game *Game) Update() error {
	if game.pollDesktopUpdates() {
		return ebiten.Termination
	}
	if game.runContext != nil && game.runContext.Err() != nil {
		return ebiten.Termination
	}
	if game.cancelInputRequested.Swap(false) {
		game.cancelPointer()
	}
	game.receivePreferences()
	game.pollProjects()
	game.pollUpdateDiscovery()
	game.pollDirectoryStatus()
	game.workspaceBackEnabled.Store(game.inWorkspace())
	if game.workspaceBackRequested.Swap(false) {
		game.workspaceBack()
	}
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
	game.clampBodyScroll()
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
		game.beginPTT("right-ctrl")
	}
	if !rightCtrl && game.rightCtrlHeld {
		game.endPTT()
	}
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		game.cancelPTT()
		game.setPTTHeld(false)
	}
	game.rightCtrlHeld = rightCtrl
	game.updateGamepad()
	game.pollSpeech()
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
			game.beginPTT("gamepad-r2")
		}
		if !r2 && game.r2Held {
			game.endPTT()
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
	if !game.inWorkspace() || game.ui.GetFocusedWidget() != nil || game.voice.canvas != nil {
		game.drawPTTButton(screen)
	}
	if game.debugOverlayOpen {
		game.drawDebugOverlay(screen)
	}
	if game.debugOverlayOpen || !game.inWorkspace() {
		drawDiagnostics(screen, game.shell.Snapshot(), game.height)
	}
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
