//go:build gui

package gui

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func defaultRuntimeInfo() RuntimeInfo {
	wd, _ := os.Getwd()
	binary, _ := os.Executable()
	return RuntimeInfo{
		WorkingDir:  wd,
		RuntimePath: os.Getenv("APPARAT_RUNTIME_DIR"),
		BinaryPath:  binary,
	}
}

func (game *Game) drawDebugOverlay(screen *ebiten.Image) {
	x, y := game.debugOverlayX, game.debugOverlayY
	width, height := game.debugOverlaySize()
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(width), float64(height), color.RGBA{R: 12, G: 16, B: 24, A: 232})
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(width), 30, activeTabColor)
	ebitenutil.DebugPrintAt(screen, "Debug UI", x+10, y+8)
	for index, line := range game.debugOverlayLines() {
		ebitenutil.DebugPrintAt(screen, line, x+10, y+42+(index*22))
	}
}

func (game *Game) debugOverlaySize() (int, int) {
	width := 520
	if game.width > 0 && game.width-24 < width {
		width = game.width - 24
	}
	if width < 260 {
		width = 260
	}
	return width, 220
}

func (game *Game) debugOverlayLines() []string {
	snapshot := game.shell.Snapshot()
	return []string{
		fmt.Sprintf("screen=%dx%d fps=%.1f ups=%.1f", game.width, game.height, ebiten.ActualFPS(), ebiten.ActualTPS()),
		fmt.Sprintf("active=%s route=%s input=%s", snapshot.ActiveTab().ID(), snapshot.Diagnostics.ActiveRoute, snapshot.Diagnostics.Input),
		fmt.Sprintf("voice=%s focus=%s queue=%d", snapshot.VoiceState, snapshot.Diagnostics.Focused, snapshot.Diagnostics.EventQueueSize),
		fmt.Sprintf("working=%s", fallback(game.runtimeInfo.WorkingDir, "unknown")),
		fmt.Sprintf("runtime=%s", fallback(game.runtimeInfo.RuntimePath, "not configured")),
		fmt.Sprintf("binary=%s", fallback(game.runtimeInfo.BinaryPath, "unknown")),
	}
}

func fallback(value string, replacement string) string {
	if value == "" {
		return replacement
	}
	return filepath.Clean(value)
}

func (game *Game) collapsed() bool {
	return game.width < collapsedWidth
}

func (game *Game) selectSection(tabID hud.TabID, index int) {
	for _, tabData := range game.shell.Snapshot().Tabs {
		if tabData.ID() == tabID && !tabData.IsSelectableSection(index) {
			return
		}
	}
	game.selectedSections[tabID] = index
	if game.collapsed() {
		game.detailOpen[tabID] = true
	}
	game.rebuildUI(game.shell.Snapshot())
}

func (game *Game) showList(tabID hud.TabID) {
	game.detailOpen[tabID] = false
	game.rebuildUI(game.shell.Snapshot())
}

func (game *Game) clampSplitWidth() {
	if game.splitWidth < minSplitWidth {
		game.splitWidth = minSplitWidth
	}
	limit := game.width - 220
	if limit > maxSplitWidth {
		limit = maxSplitWidth
	}
	if limit < minSplitWidth {
		limit = minSplitWidth
	}
	if game.splitWidth > limit {
		game.splitWidth = limit
	}
}

func (game *Game) ensureActiveTabVisible() {
	if game.tabScroll == nil || game.tabButtonCount <= 1 || !game.activeTabScrollPending {
		return
	}
	game.activeTabScrollPending = false
	if !game.tabStripOverflows() {
		game.tabScroll.ScrollLeft = 0
		game.tabStripScrollLeft = 0
		return
	}
	active := game.shell.Snapshot().ActiveIndex
	if active < 0 || active >= len(game.tabButtonLefts) {
		return
	}
	viewportWidth := game.hudPreferredWidth()
	maxOffset := game.tabStripContentWidth - viewportWidth
	if maxOffset <= 0 {
		return
	}
	offset := game.tabStripScrollLeft * float64(maxOffset)
	left := float64(game.tabButtonLefts[active])
	right := float64(game.tabButtonRights[active])
	switch {
	case left < offset:
		offset = left
	case right > offset+float64(viewportWidth):
		offset = right - float64(viewportWidth)
	}
	offset = float64(max(0, min(int(offset), maxOffset)))
	game.tabStripScrollLeft = float64(offset) / float64(maxOffset)
	game.tabScroll.ScrollLeft = game.tabStripScrollLeft
}

func (game *Game) requestActiveTabVisible() {
	game.activeTabScrollPending = true
}

func (game *Game) updatePointerState() {
	x, y := ebiten.CursorPosition()
	game.updateDebugOverlayDrag(x, y)
	game.updateTabStripDrag(x, y)
	game.updateBodyScrollMouse(x, y)
	game.updatePTTMouse(x, y)
	game.updateTouchDrag()
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		game.splitDragging = false
	}
}

func (game *Game) updateDebugOverlayDrag(x int, y int) {
	if !game.debugOverlayOpen {
		game.debugDrag = false
		return
	}
	width, _ := game.debugOverlaySize()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && pointInRect(x, y, game.debugOverlayX, game.debugOverlayY, width, 30) {
		game.debugDrag = true
		game.debugDragDX = x - game.debugOverlayX
		game.debugDragDY = y - game.debugOverlayY
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		game.debugDrag = false
	}
	if game.debugDrag {
		game.clampDebugOverlay(x-game.debugDragDX, y-game.debugDragDY)
	}
}

func (game *Game) updateTouchDrag() {
	for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
		x, y := ebiten.TouchPosition(id)
		width, _ := game.debugOverlaySize()
		if game.debugOverlayOpen && pointInRect(x, y, game.debugOverlayX, game.debugOverlayY, width, 30) {
			game.debugTouchActive = true
			game.debugTouchID = id
			game.debugDragDX = x - game.debugOverlayX
			game.debugDragDY = y - game.debugOverlayY
			return
		}
		if game.tabScroll != nil && imagePoint(x, y).In(game.tabScroll.GetWidget().Rect) {
			game.tabTouchActive = true
			game.tabTouchID = id
			game.tabStripLastX = x
			game.tabStripDragDistance = 0
			game.tabStripDragMoved = false
			return
		}
		if !game.pointInPTT(x, y) && !game.pointInDebugOverlay(x, y) && game.beginBodyTouch(id, x, y) {
			return
		}
		if game.pointInPTT(x, y) {
			game.pttTouchActive = true
			game.pttTouchID = id
			game.setPTTHeld(true)
			return
		}
	}
	if game.debugTouchActive {
		game.updateDebugTouch()
	}
	if game.tabTouchActive {
		game.updateTabTouch()
	}
	if game.bodyTouchActive {
		game.updateBodyTouch()
	}
	if game.pttTouchActive {
		game.updatePTTTouch()
	}
}

func (game *Game) updateDebugTouch() {
	if inpututil.IsTouchJustReleased(game.debugTouchID) {
		game.debugTouchActive = false
		return
	}
	x, y := ebiten.TouchPosition(game.debugTouchID)
	game.clampDebugOverlay(x-game.debugDragDX, y-game.debugDragDY)
}

func (game *Game) clampDebugOverlay(x int, y int) {
	width, height := game.debugOverlaySize()
	game.debugOverlayX = clamp(x, 0, max(0, game.width-width))
	game.debugOverlayY = clamp(y, 0, max(0, game.height-height))
}

func (game *Game) pointInDebugOverlay(x int, y int) bool {
	if !game.debugOverlayOpen {
		return false
	}
	width, height := game.debugOverlaySize()
	return pointInRect(x, y, game.debugOverlayX, game.debugOverlayY, width, height)
}

func (game *Game) updateTabTouch() {
	if inpututil.IsTouchJustReleased(game.tabTouchID) || game.tabScroll == nil {
		game.finishTabStripDrag()
		game.tabTouchActive = false
		return
	}
	x, _ := ebiten.TouchPosition(game.tabTouchID)
	dx := x - game.tabStripLastX
	game.tabStripLastX = x
	game.scrollTabStripBy(dx)
}

func (game *Game) updateTabStripDrag(x int, y int) {
	if game.tabScroll == nil {
		return
	}
	rect := game.tabScroll.GetWidget().Rect
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && imagePoint(x, y).In(rect) {
		game.tabStripDragging = true
		game.tabStripLastX = x
		game.tabStripDragDistance = 0
		game.tabStripDragMoved = false
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		game.finishTabStripDrag()
		game.tabStripDragging = false
	}
	if game.tabStripDragging && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		dx := x - game.tabStripLastX
		game.tabStripLastX = x
		game.scrollTabStripBy(dx)
	}
	if imagePoint(x, y).In(rect) {
		wheelX, wheelY := ebiten.Wheel()
		if wheelX != 0 || wheelY != 0 {
			game.tabScroll.ScrollLeft -= (wheelX + wheelY) / 12
		}
	}
}

func (game *Game) trackTabStripDrag(dx int) {
	if dx < 0 {
		dx = -dx
	}
	game.tabStripDragDistance += dx
	if game.tabStripDragDistance >= tabDragThreshold {
		game.tabStripDragMoved = true
	}
}

func (game *Game) scrollTabStripBy(dx int) {
	if game.tabScroll == nil || !game.tabStripOverflows() {
		return
	}
	game.trackTabStripDrag(dx)
	game.tabScroll.ScrollLeft -= float64(dx) / 360.0
	game.tabStripScrollLeft = clampScrollLeft(game.tabScroll.ScrollLeft)
	game.tabScroll.ScrollLeft = game.tabStripScrollLeft
}

func (game *Game) tabStripOverflows() bool {
	return game.tabStripContentWidth > game.hudPreferredWidth()
}

func clampScrollLeft(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func (game *Game) finishTabStripDrag() {
	if game.tabStripDragMoved {
		game.tabDragCancelUpdates = 2
		game.syncTabButtonStates()
	}
	game.tabStripDragDistance = 0
	game.tabStripDragMoved = false
}

func (game *Game) tabSelectionSuppressed() bool {
	return game.tabStripDragMoved || game.tabDragCancelUpdates > 0
}

func (game *Game) advanceTabDragCancellation() {
	if game.tabDragCancelUpdates > 0 {
		game.tabDragCancelUpdates--
	}
}

func (game *Game) updatePTTMouse(x int, y int) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && game.pointInPTT(x, y) {
		game.setPTTHeld(true)
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && game.pttHeld && !game.pttTouchActive {
		game.setPTTHeld(false)
	}
}

func (game *Game) updatePTTTouch() {
	if inpututil.IsTouchJustReleased(game.pttTouchID) {
		game.pttTouchActive = false
		game.setPTTHeld(false)
	}
}

func pointInRect(px int, py int, x int, y int, w int, h int) bool {
	return px >= x && px <= x+w && py >= y && py <= y+h
}

func imagePoint(x int, y int) image.Point {
	return image.Point{X: x, Y: y}
}

func clamp(value int, low int, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
