//go:build gui

package gui

import (
	"fmt"
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
	tabTextInsetX     = 12
	tabTextInsetY     = 18
	tabWidth          = 132
	tabWideWidth      = 156
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
}

type tabRect struct {
	index int
	x     int
	y     int
	w     int
	h     int
}

func NewGame() *Game {
	root := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewAnchorLayout()))
	return &Game{shell: hud.NewShell(), ui: &ebitenui.UI{Container: root}, width: 1280, height: 800}
}

func (game *Game) Update() error {
	game.ui.Update()
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
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if index, ok := game.tabIndexAt(x, y); ok {
			_ = game.shell.SelectTab(index)
		}
	}
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
	return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
	snapshot := game.shell.Snapshot()
	screen.Fill(backgroundColor)
	game.ui.Draw(screen)
	game.tabRects = drawTabs(screen, snapshot)
	drawActiveTab(screen, snapshot, game.width, game.height)
	drawDiagnostics(screen, snapshot, game.height)
}

func drawTabs(screen *ebiten.Image, snapshot hud.Snapshot) []tabRect {
	rects := make([]tabRect, 0, len(snapshot.Tabs))
	x := windowMargin
	for index, tab := range snapshot.Tabs {
		label := fmt.Sprintf("%s %s", tab.Descriptor.Glyph, tab.Title())
		w := tabWidth
		if len(label) > 11 {
			w = tabWideWidth
		}
		color := panelColor
		if index == snapshot.ActiveIndex {
			color = activeTabColor
		}
		ebitenutil.DrawRect(screen, float64(x), tabTop, float64(w), tabHeight, color)
		ebitenutil.DebugPrintAt(screen, label, x+tabTextInsetX, tabTop+tabTextInsetY)
		rects = append(rects, tabRect{index: index, x: x, y: tabTop, w: w, h: tabHeight})
		x += w + tabGap
	}
	return rects
}

func (game *Game) tabIndexAt(x int, y int) (int, bool) {
	for _, rect := range game.tabRects {
		if x >= rect.x && x < rect.x+rect.w && y >= rect.y && y < rect.y+rect.h {
			return rect.index, true
		}
	}
	return 0, false
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
