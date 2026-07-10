//go:build gui

package gui

import (
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	bodyDragThreshold = 8
	scrollWheelStep   = 42
)

type scrollPane string

const (
	scrollPaneSettings scrollPane = "settings"
	scrollPaneList     scrollPane = "list"
	scrollPaneDetail   scrollPane = "detail"
)

type bodyScrollState struct {
	settings int
	list     int
	detail   int
}

type bodyDragState struct {
	active      bool
	pane        scrollPane
	dragged     bool
	startY      int
	startScroll int
}

func (game *Game) updateBodyWheel() {
	_, wheelY := ebiten.Wheel()
	if wheelY == 0 {
		return
	}
	x, y := ebiten.CursorPosition()
	pane, ok := game.scrollPaneAt(x, y)
	if !ok {
		return
	}
	game.scrollBodyPane(pane, int(-wheelY*scrollWheelStep))
}

func (game *Game) updateMouseBodyDrag() {
	x, y := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && !game.pointInTabStrip(x, y) {
		if pane, ok := game.scrollPaneAt(x, y); ok {
			game.mouseBodyDrag = bodyDragState{active: true, pane: pane, startY: y, startScroll: game.scrollForPane(pane)}
		}
	}
	if game.mouseBodyDrag.active && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		game.dragBodyPane(&game.mouseBodyDrag, y)
	}
	if game.mouseBodyDrag.active && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		game.mouseBodyDrag = bodyDragState{}
	}
}

func (game *Game) updateTouchBodyDrag() {
	if !game.touchBodyDrag.active {
		for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
			x, y := ebiten.TouchPosition(id)
			if game.pointInTabStrip(x, y) {
				continue
			}
			pane, ok := game.scrollPaneAt(x, y)
			if !ok {
				continue
			}
			game.bodyTouchID = id
			game.touchBodyDrag = bodyDragState{active: true, pane: pane, startY: y, startScroll: game.scrollForPane(pane)}
			break
		}
	}
	if !game.touchBodyDrag.active {
		return
	}
	_, y := ebiten.TouchPosition(game.bodyTouchID)
	game.dragBodyPane(&game.touchBodyDrag, y)
	for _, id := range inpututil.AppendJustReleasedTouchIDs(nil) {
		if id == game.bodyTouchID {
			game.touchBodyDrag = bodyDragState{}
			break
		}
	}
}

func (game *Game) dragBodyPane(state *bodyDragState, y int) {
	dy := y - state.startY
	if !state.dragged && absInt(dy) > bodyDragThreshold {
		state.dragged = true
		state.startY = y
		state.startScroll = game.scrollForPane(state.pane)
		dy = 0
	}
	if state.dragged {
		game.setScrollForPane(state.pane, state.startScroll-dy)
	}
}

func (game *Game) scrollPaneAt(x int, y int) (scrollPane, bool) {
	body := tabBodyRect(game.width, game.height)
	if !body.contains(x, y) {
		return "", false
	}
	tab := game.shell.Snapshot().ActiveTab()
	if tab.ID() == hud.TabSettings {
		return scrollPaneSettings, true
	}
	list, detail := masterDetailRects(body)
	if list.contains(x, y) {
		return scrollPaneList, true
	}
	if detail.contains(x, y) {
		return scrollPaneDetail, true
	}
	return "", false
}

func (game *Game) scrollBodyPane(pane scrollPane, delta int) {
	game.setScrollForPane(pane, game.scrollForPane(pane)+delta)
}

func (game *Game) scrollForPane(pane scrollPane) int {
	switch pane {
	case scrollPaneSettings:
		return game.bodyScroll.settings
	case scrollPaneList:
		return game.bodyScroll.list
	case scrollPaneDetail:
		return game.bodyScroll.detail
	default:
		return 0
	}
}

func (game *Game) setScrollForPane(pane scrollPane, scroll int) {
	max := game.maxScrollForPane(pane)
	scroll = clampScroll(scroll, max)
	switch pane {
	case scrollPaneSettings:
		game.bodyScroll.settings = scroll
	case scrollPaneList:
		game.bodyScroll.list = scroll
	case scrollPaneDetail:
		game.bodyScroll.detail = scroll
	}
}

func (game *Game) maxScrollForPane(pane scrollPane) int {
	body := tabBodyRect(game.width, game.height)
	tab := game.shell.Snapshot().ActiveTab()
	switch pane {
	case scrollPaneSettings:
		return settingsContentHeight(tab) - body.h
	case scrollPaneList:
		list, _ := masterDetailRects(body)
		return listContentHeight(tab) - list.h
	case scrollPaneDetail:
		_, detail := masterDetailRects(body)
		return detailContentHeight(tab) - detail.h
	default:
		return 0
	}
}

func clampScroll(scroll int, max int) int {
	if max < 0 {
		max = 0
	}
	if scroll < 0 {
		return 0
	}
	if scroll > max {
		return max
	}
	return scroll
}
