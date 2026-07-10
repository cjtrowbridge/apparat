//go:build gui

package gui

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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
	tabDragThreshold  = 6
	bodyGap           = 8
	bodyInnerPaddingX = 18
	bodyInnerPaddingY = 18
	diagnosticsHeight = 58
)

type Game struct {
	shell          hud.Shell
	ui             *ebitenui.UI
	width          int
	height         int
	rightCtrlHeld  bool
	pageDownWasHit bool
	pageUpWasHit   bool
	tabRects       []tabRect
	tabScrollX     int
	tabContentW    int
	mouseTabDrag   dragState
	touchTabDrag   dragState
	touchID        ebiten.TouchID
	l1WasPressed   bool
	r1WasPressed   bool
	r2Held         bool
}

type tabRect struct {
	index int
	x     int
	y     int
	w     int
	h     int
}

type dragState struct {
	active      bool
	dragged     bool
	startX      int
	lastX       int
	startY      int
	startScroll int
}

func NewGame() *Game {
	root := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewAnchorLayout()))
	return &Game{shell: hud.NewShell(), ui: &ebitenui.UI{Container: root}, width: 1280, height: 800}
}

func (game *Game) Update() error {
	game.ui.Update()
	startIndex := game.shell.Snapshot().ActiveIndex
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
	game.updateMouseTabDrag()
	game.updateTouchTabDrag()
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
		game.ensureTabVisible(activeIndex)
	}
	return nil
}

func (game *Game) updateMouseTabDrag() {
	x, y := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && game.pointInTabStrip(x, y) {
		game.mouseTabDrag = dragState{active: true, startX: x, lastX: x, startY: y, startScroll: game.tabScrollX}
	}
	if game.mouseTabDrag.active && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		game.dragTabs(&game.mouseTabDrag, x)
	}
	if game.mouseTabDrag.active && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		if !game.mouseTabDrag.dragged {
			if index, ok := game.tabIndexAt(x, y); ok {
				_ = game.shell.SelectTab(index)
			}
		}
		game.mouseTabDrag = dragState{}
	}
}

func (game *Game) updateTouchTabDrag() {
	if !game.touchTabDrag.active {
		for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
			x, y := ebiten.TouchPosition(id)
			if game.pointInTabStrip(x, y) {
				game.touchID = id
				game.touchTabDrag = dragState{active: true, startX: x, lastX: x, startY: y, startScroll: game.tabScrollX}
				break
			}
		}
	}
	if game.touchTabDrag.active {
		x, _ := ebiten.TouchPosition(game.touchID)
		game.dragTabs(&game.touchTabDrag, x)
		for _, id := range inpututil.AppendJustReleasedTouchIDs(nil) {
			if id != game.touchID {
				continue
			}
			if !game.touchTabDrag.dragged {
				if index, ok := game.tabIndexAt(game.touchTabDrag.startX, game.touchTabDrag.startY); ok {
					_ = game.shell.SelectTab(index)
				}
			}
			game.touchTabDrag = dragState{}
			break
		}
	}
}

func (game *Game) dragTabs(state *dragState, x int) {
	dx := x - state.startX
	if dx < -tabDragThreshold || dx > tabDragThreshold {
		state.dragged = true
	}
	if state.dragged {
		game.tabScrollX = clampTabScroll(state.startScroll-dx, game.tabContentW, game.tabViewportWidth())
	}
	state.lastX = x
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
	snapshot := game.shell.Snapshot()
	screen.Fill(backgroundColor)
	game.ui.Draw(screen)
	game.tabRects = game.drawTabs(screen, snapshot)
	drawActiveTab(screen, snapshot, game.width, game.height)
	drawDiagnostics(screen, snapshot, game.height)
}

func (game *Game) drawTabs(screen *ebiten.Image, snapshot hud.Snapshot) []tabRect {
	rects := make([]tabRect, 0, len(snapshot.Tabs))
	tabW := tabButtonWidth(snapshot)
	game.tabContentW = tabContentWidth(len(snapshot.Tabs), tabW)
	game.tabScrollX = clampTabScroll(game.tabScrollX, game.tabContentW, game.tabViewportWidth())
	tabArea := screen.SubImage(image.Rect(windowMargin, tabTop, windowMargin+game.tabViewportWidth(), tabTop+tabHeight)).(*ebiten.Image)
	x := -game.tabScrollX
	for index, tab := range snapshot.Tabs {
		label := fmt.Sprintf("%s %s", tab.Descriptor.Glyph, tab.Title())
		color := panelColor
		if index == snapshot.ActiveIndex {
			color = activeTabColor
		}
		ebitenutil.DrawRect(tabArea, float64(x), 0, float64(tabW), tabHeight, color)
		ebitenutil.DebugPrintAt(tabArea, label, x+tabTextInsetX, tabTextInsetY)
		rects = append(rects, tabRect{index: index, x: windowMargin + x, y: tabTop, w: tabW, h: tabHeight})
		x += tabW + tabGap
	}
	return rects
}

func tabButtonWidth(snapshot hud.Snapshot) int {
	maxW := 0
	for _, tab := range snapshot.Tabs {
		label := fmt.Sprintf("%s %s", tab.Descriptor.Glyph, tab.Title())
		if w := labelWidth(label); w > maxW {
			maxW = w
		}
	}
	return maxW + tabTextInsetX*2
}

func labelWidth(label string) int {
	return len([]rune(label)) * debugGlyphWidth
}

func tabContentWidth(count int, tabW int) int {
	if count == 0 {
		return 0
	}
	return count*tabW + (count-1)*tabGap
}

func (game *Game) tabIndexAt(x int, y int) (int, bool) {
	if !game.pointInTabStrip(x, y) {
		return 0, false
	}
	for _, rect := range game.tabRects {
		if x >= rect.x && x < rect.x+rect.w && y >= rect.y && y < rect.y+rect.h {
			return rect.index, true
		}
	}
	return 0, false
}

func (game *Game) pointInTabStrip(x int, y int) bool {
	return x >= windowMargin && x < game.width-windowMargin && y >= tabTop && y < tabTop+tabHeight
}

func (game *Game) tabViewportWidth() int {
	if width := game.width - windowMargin*2; width > 0 {
		return width
	}
	return 1
}

func (game *Game) ensureTabVisible(index int) {
	snapshot := game.shell.Snapshot()
	tabW := tabButtonWidth(snapshot)
	game.tabContentW = tabContentWidth(len(snapshot.Tabs), tabW)
	viewportW := game.tabViewportWidth()
	left := index * (tabW + tabGap)
	right := left + tabW
	if left < game.tabScrollX {
		game.tabScrollX = left
	}
	if right > game.tabScrollX+viewportW {
		game.tabScrollX = right - viewportW
	}
	game.tabScrollX = clampTabScroll(game.tabScrollX, game.tabContentW, viewportW)
}

func clampTabScroll(scroll int, contentW int, viewportW int) int {
	maxScroll := contentW - viewportW
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scroll < 0 {
		return 0
	}
	if scroll > maxScroll {
		return maxScroll
	}
	return scroll
}

func drawActiveTab(screen *ebiten.Image, snapshot hud.Snapshot, width int, height int) {
	tab := snapshot.ActiveTab()
	bodyX := windowMargin
	bodyY := tabTop + tabHeight + bodyGap
	bodyWidth := width - windowMargin*2
	bodyHeight := height - bodyY - diagnosticsHeight
	contentX := bodyX + bodyInnerPaddingX
	contentY := bodyY + bodyInnerPaddingY
	ebitenutil.DrawRect(screen, float64(bodyX), float64(bodyY), float64(bodyWidth), float64(bodyHeight), panelColor)
	ebitenutil.DebugPrintAt(screen, tab.Title(), contentX, contentY)
	ebitenutil.DebugPrintAt(screen, tab.Summary, contentX, contentY+24)
	y := contentY + 60
	for _, section := range tab.Sections {
		ebitenutil.DebugPrintAt(screen, strings.ToUpper(section.Title), contentX, y)
		y += 24
		for _, row := range section.Rows {
			prefix := "•"
			if row.Disabled {
				prefix = "⊘"
			} else if row.Future {
				prefix = "◇"
			}
			line := fmt.Sprintf("%s %s", prefix, row.Label)
			if row.Detail != "" {
				line = fmt.Sprintf("%s — %s", line, row.Detail)
			}
			ebitenutil.DebugPrintAt(screen, line, contentX+20, y)
			y += 20
		}
		y += 12
	}
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
		game.width = outsideWidth
	}
	if outsideHeight > 0 {
		game.height = outsideHeight
	}
	return game.width, game.height
}
