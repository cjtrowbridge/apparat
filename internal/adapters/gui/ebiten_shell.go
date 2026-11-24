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
	actions                        actionScreen
	nav                            navigation
	voice                          voiceState
	globalDraft                    string
	globalVoiceFocus               bool
	globalField                    *widget.TextInput
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
	formLabelFace                  *ebitentext.Face
	formDescriptionFace            *ebitentext.Face
	width                          int
	height                         int
	layoutDirty                    bool
	rightCtrlHeld                  bool
	pageDownWasHit                 bool
	pageUpWasHit                   bool
	l1WasPressed                   bool
	r1WasPressed                   bool
	l2Held                         bool
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
	cluster                        clusterNavigation
	lan                            lanView
	runtimeInfo                    RuntimeInfo
	activeTabID                    atomic.Value
	updateStatus                   atomic.Value
	updateButton                   *widget.Button
	updateButtonStatus             string
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
	game.leaveGlobalPrompt()
	if ebiten.IsFocused() {
		game.receiveNativePrompt()
	}
	game.receivePreferences()
	game.pollProjects()
	game.pollUpdateDiscovery()
	game.openRequestedDiscovery()
	game.pollLAN()
	game.pollCluster()
	clusterOpen := game.shell.Snapshot().ActiveTab().ID() == hud.TabCluster && game.clusterTop() != nil
	game.pollDirectoryStatus()
	peerOpen := game.shell.Snapshot().ActiveTab().ID() == hud.TabCluster && game.lan.selected != ""
	game.workspaceBackEnabled.Store(game.actions.mode != "" || game.inWorkspace() || peerOpen || clusterOpen)
	if game.workspaceBackRequested.Swap(false) {
		if game.actions.mode != "" {
			game.backActionScreen()
		} else if clusterOpen {
			game.clusterBack()
		} else if peerOpen {
			game.leavePeer()
		} else {
			game.workspaceBack()
		}
	}
	game.trackWindow()
	game.receiveServices()
	if game.layoutDirty && game.gesture.Owner == gestureNone {
		game.layoutDirty = false
		game.rebuildUI(game.shell.Snapshot())
	}
	if game.updatePTTPointer() {
		game.keyboardPTT()
		game.pollSpeech()
		return nil
	}
	if game.actions.mode == "" && game.arcadeUpdate() {
		game.keyboardPTT()
		game.arcadeTriggerPTT()
		game.pollSpeech()
		return nil
	}
	game.globalTypingPress()
	game.cancelSpeechBeforeEditing()
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

	for index, key := range []ebiten.Key{ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4, ebiten.Key5, ebiten.Key6, ebiten.Key7} {
		if ebiten.IsKeyPressed(ebiten.KeyAlt) && ebiten.IsKeyPressed(key) {
			game.selectTabWithReveal(index)
		}
	}
	game.keyboardPTT()
	game.navigateInputs()
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
		if game.desktopUpdates != nil {
			game.updateButton.GetWidget().Disabled = game.desktopUpdates.busy
		} else {
			game.updateButton.GetWidget().Disabled = game.androidTransferBusy() || game.discovery.pending != nil
		}
		if game.updateButtonStatus != status {
			game.updateButtonStatus = status
			game.updateButton.SetText(game.updateLabel(status))
			game.ui.Container.RequestRelayout()
		}
	}
	if game.updateDetail != nil {
		detail := game.UpdateStatusDetail()
		if detail != game.updateDetail.Label {
			game.updateDetail.Label = detail
			game.ui.Container.RequestRelayout()
		}
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

		l1 := ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonFrontTopLeft)
		r1 := ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonFrontTopRight)
		l2 := ebiten.IsStandardGamepadButtonPressed(id, pttTriggerButton)
		if r1 && !game.r1WasPressed {
			game.applyTabAction(hud.ActionNextTab)
		}
		if l1 && !game.l1WasPressed {
			game.applyTabAction(hud.ActionPreviousTab)
		}
		game.setTriggerPTT(l2)
		game.l1WasPressed = l1
		game.r1WasPressed = r1
		return // Use only the first standard-layout gamepad.
	}
	game.cancelTriggerPTT()
	game.l1WasPressed = false
	game.r1WasPressed = false
}

func (game *Game) applyTabAction(action hud.Action) {
	if game.actions.mode != "" {
		return
	}
	active := game.shell.Snapshot().ActiveIndex
	{
		err := game.shell.ApplyAction(action)
		if err == nil && game.shell.Snapshot().ActiveIndex != active {
			game.requestActiveTabVisible()
		}
	}
}

func (game *Game) selectTabWithReveal(index int) {
	if game.actions.mode != "" {
		return
	}
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
	bg := game.actions.background
	if bg != nil {
		bg.Draw(screen)
		if !game.actions.returnToBrowser {
			game.arcadeDraw(screen)
		}
	}
	if game.ui != nil {
		game.ui.Draw(screen)
		game.revealSetting()
	}
	if game.actions.mode == "" && game.arcadeDraw(screen) {
		game.drawArcadeFocus(screen)
		game.drawPTTButton(screen)
		return
	}
	game.drawFocus(screen)
	if game.actions.mode != "" {
		game.drawPTTButton(screen)
		return
	}
	if game.debugOverlayOpen {
		game.drawDebugOverlay(screen)
	}
	if game.debugOverlayOpen || !game.inWorkspace() {
		drawDiagnostics(screen, game.shell.Snapshot(), game.height)
	}
	game.drawPTTButton(screen)
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
