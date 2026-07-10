//go:build gui

package gui

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	bodyBorderColor   = color.RGBA{R: 58, G: 70, B: 96, A: 255}
	fieldsetColor     = color.RGBA{R: 22, G: 27, B: 42, A: 255}
	listPaneColor     = color.RGBA{R: 20, G: 24, B: 38, A: 255}
	selectedItemColor = color.RGBA{R: 45, G: 64, B: 98, A: 255}
)

const (
	bodyInset          = 18
	bodySectionGap     = 12
	fieldsetPadding    = 12
	fieldsetTitleH     = 20
	fieldsetDescH      = 18
	touchTargetH       = 44
	fieldsetRowH       = touchTargetH
	fieldsetMinH       = 112
	masterMinListW     = 170
	masterListRatioNum = 3
	masterListRatioDen = 10
	masterDividerW     = 8
	nativeSlotUpdate   = "settings.updates.check_for_update"
)

type rect struct {
	x int
	y int
	w int
	h int
}

func (area rect) contains(x int, y int) bool {
	return x >= area.x && x < area.x+area.w && y >= area.y && y < area.y+area.h
}

func (game *Game) drawActiveTab(screen *ebiten.Image, snapshot hud.Snapshot) {
	body := tabBodyRect(game.width, game.height)
	ebitenutil.DrawRect(screen, float64(body.x), float64(body.y), float64(body.w), float64(body.h), panelColor)
	tab := snapshot.ActiveTab()
	if tab.ID() == hud.TabSettings {
		game.drawSettingsBody(screen, tab, body)
		return
	}
	game.drawMasterDetailBody(screen, tab, body)
}

func tabBodyRect(width int, height int) rect {
	y := tabTop + tabHeight + bodyGap
	return rect{x: windowMargin, y: y, w: width - windowMargin*2, h: height - y - diagnosticsHeight}
}

func (game *Game) drawSettingsBody(screen *ebiten.Image, tab hud.Tab, body rect) {
	game.bodyScroll.settings = clampScroll(game.bodyScroll.settings, settingsContentHeight(tab)-body.h)
	pane := clippedImage(screen, body)
	x := body.x + bodyInset
	y := body.y + bodyInset - game.bodyScroll.settings
	w := body.w - bodyInset*2
	ebitenutil.DebugPrintAt(pane, tab.Title(), x-body.x, y-body.y)
	drawTextBlock(pane, tab.Summary, rect{x: x - body.x, y: y + 24 - body.y, w: w, h: fieldsetDescH * 2})
	y += 58
	for _, section := range tab.Sections {
		h := fieldsetHeight(section)
		drawFieldset(pane, rect{x: x - body.x, y: y - body.y, w: w, h: h}, section)
		y += h + bodySectionGap
	}
}

func (game *Game) drawMasterDetailBody(screen *ebiten.Image, tab hud.Tab, body rect) {
	list, detail := masterDetailRects(body)
	game.bodyScroll.list = clampScroll(game.bodyScroll.list, listContentHeight(tab)-list.h)
	game.bodyScroll.detail = clampScroll(game.bodyScroll.detail, detailContentHeight(tab)-detail.h)
	drawListPane(screen, tab, list, game.bodyScroll.list)
	drawDetailPane(screen, tab, detail, game.bodyScroll.detail)
}

func masterDetailRects(body rect) (rect, rect) {
	inner := rect{x: body.x + bodyInset, y: body.y + bodyInset, w: body.w - bodyInset*2, h: body.h - bodyInset*2}
	listW := inner.w * masterListRatioNum / masterListRatioDen
	if listW < masterMinListW {
		listW = masterMinListW
	}
	maxListW := inner.w - masterDividerW - masterMinListW
	if maxListW < masterMinListW {
		maxListW = inner.w / 2
	}
	if listW > maxListW {
		listW = maxListW
	}
	list := rect{x: inner.x, y: inner.y, w: listW, h: inner.h}
	detail := rect{x: inner.x + listW + masterDividerW, y: inner.y, w: inner.w - listW - masterDividerW, h: inner.h}
	return list, detail
}

func drawListPane(screen *ebiten.Image, tab hud.Tab, pane rect, scroll int) {
	ebitenutil.DrawRect(screen, float64(pane.x), float64(pane.y), float64(pane.w), float64(pane.h), listPaneColor)
	drawBorder(screen, pane)
	clipped := clippedImage(screen, pane)
	ebitenutil.DebugPrintAt(clipped, tab.Title(), fieldsetPadding, fieldsetPadding-scroll)
	y := fieldsetPadding + 28 - scroll
	items := listItemsForTab(tab)
	for index, item := range items {
		if index == 0 {
			ebitenutil.DrawRect(clipped, 6, float64(y-6), float64(pane.w-12), float64(touchTargetH), selectedItemColor)
		}
		ebitenutil.DebugPrintAt(clipped, truncateText(item, pane.w-fieldsetPadding*2), fieldsetPadding, y)
		y += touchTargetH
	}
}

func drawDetailPane(screen *ebiten.Image, tab hud.Tab, pane rect, scroll int) {
	ebitenutil.DrawRect(screen, float64(pane.x), float64(pane.y), float64(pane.w), float64(pane.h), fieldsetColor)
	drawBorder(screen, pane)
	clipped := clippedImage(screen, pane)
	x := fieldsetPadding
	y := fieldsetPadding - scroll
	ebitenutil.DebugPrintAt(clipped, "Placeholder Detail", x, y)
	drawTextBlock(clipped, tab.Summary, rect{x: x, y: y + 24, w: pane.w - fieldsetPadding*2, h: fieldsetDescH * 2})
	y += 58
	for _, section := range tab.Sections {
		h := fieldsetHeight(section)
		drawFieldset(clipped, rect{x: x, y: y, w: pane.w - fieldsetPadding*2, h: h}, section)
		y += h + bodySectionGap
	}
}

func drawFieldset(screen *ebiten.Image, area rect, section hud.Section) {
	ebitenutil.DrawRect(screen, float64(area.x), float64(area.y), float64(area.w), float64(area.h), fieldsetColor)
	drawBorder(screen, area)
	x := area.x + fieldsetPadding
	y := area.y + fieldsetPadding
	ebitenutil.DebugPrintAt(screen, strings.ToUpper(section.Title), x, y)
	y += fieldsetTitleH
	if section.Description != "" {
		lines := drawTextBlock(screen, section.Description, rect{x: x, y: y, w: area.w - fieldsetPadding*2, h: fieldsetDescH * 2})
		y += lines * fieldsetDescH
	}
	for _, row := range section.Rows {
		ebitenutil.DebugPrintAt(screen, truncateText(rowLine(row), area.w-fieldsetPadding*2), x, y)
		y += fieldsetRowH
		if y > area.y+area.h-fieldsetRowH {
			break
		}
	}
}

func fieldsetHeight(section hud.Section) int {
	h := fieldsetPadding*2 + fieldsetTitleH + len(section.Rows)*fieldsetRowH
	if section.Description != "" {
		h += len(wrapText(section.Description, 360)) * fieldsetDescH
	}
	if h < fieldsetMinH {
		return fieldsetMinH
	}
	return h
}

func rowLine(row hud.Row) string {
	prefix := "-"
	if row.Disabled {
		prefix = "x"
	} else if row.Future {
		prefix = ">"
	}
	if row.Detail == "" {
		return fmt.Sprintf("%s %s", prefix, row.Label)
	}
	return fmt.Sprintf("%s %s: %s", prefix, row.Label, row.Detail)
}

func blockControlRect(container rect, row int) rect {
	y := container.y + fieldsetPadding + fieldsetTitleH + row*touchTargetH
	return rect{x: container.x + fieldsetPadding, y: y, w: container.w - fieldsetPadding*2, h: touchTargetH}
}

func inputPlaceholderRect(container rect, row int) rect {
	control := blockControlRect(container, row)
	if control.h < touchTargetH {
		control.h = touchTargetH
	}
	return control
}

func listItemsForTab(tab hud.Tab) []string {
	if len(tab.Sections) == 0 {
		return []string{"No items yet"}
	}
	items := make([]string, 0, len(tab.Sections))
	for _, section := range tab.Sections {
		items = append(items, section.Title)
	}
	return items
}

func truncateText(text string, maxWidth int) string {
	maxRunes := maxWidth / debugGlyphWidth
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-3]) + "..."
}

func wrapText(text string, maxWidth int) []string {
	maxRunes := maxWidth / debugGlyphWidth
	if maxRunes <= 0 {
		return []string{""}
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}
	lines := []string{}
	current := words[0]
	for _, word := range words[1:] {
		if len([]rune(current))+1+len([]rune(word)) <= maxRunes {
			current += " " + word
			continue
		}
		lines = append(lines, current)
		current = word
	}
	lines = append(lines, current)
	return lines
}

func drawTextBlock(screen *ebiten.Image, text string, area rect) int {
	lines := wrapText(text, area.w)
	maxLines := area.h / fieldsetDescH
	if maxLines <= 0 {
		return 0
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	for index, line := range lines {
		ebitenutil.DebugPrintAt(screen, truncateText(line, area.w), area.x, area.y+index*fieldsetDescH)
	}
	return len(lines)
}

func clippedImage(screen *ebiten.Image, area rect) *ebiten.Image {
	return screen.SubImage(image.Rect(area.x, area.y, area.x+area.w, area.y+area.h)).(*ebiten.Image)
}

func settingsContentHeight(tab hud.Tab) int {
	h := bodyInset + 58
	for _, section := range tab.Sections {
		h += fieldsetHeight(section) + bodySectionGap
	}
	return h + bodyInset
}

func listContentHeight(tab hud.Tab) int {
	return fieldsetPadding + 28 + len(listItemsForTab(tab))*touchTargetH + fieldsetPadding
}

func detailContentHeight(tab hud.Tab) int {
	h := fieldsetPadding + 58
	for _, section := range tab.Sections {
		h += fieldsetHeight(section) + bodySectionGap
	}
	return h + fieldsetPadding
}

func drawBorder(screen *ebiten.Image, area rect) {
	ebitenutil.DrawLine(screen, float64(area.x), float64(area.y), float64(area.x+area.w), float64(area.y), bodyBorderColor)
	ebitenutil.DrawLine(screen, float64(area.x), float64(area.y+area.h), float64(area.x+area.w), float64(area.y+area.h), bodyBorderColor)
	ebitenutil.DrawLine(screen, float64(area.x), float64(area.y), float64(area.x), float64(area.y+area.h), bodyBorderColor)
	ebitenutil.DrawLine(screen, float64(area.x+area.w), float64(area.y), float64(area.x+area.w), float64(area.y+area.h), bodyBorderColor)
}

func nativeControlSlotRect(id string, width int, height int) (rect, bool) {
	if id != nativeSlotUpdate {
		return rect{}, false
	}
	body := tabBodyRect(width, height)
	y := body.y + bodyInset + 58
	sections := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[6].Sections
	for _, section := range sections {
		if section.Title == "Updates" {
			break
		}
		y += fieldsetHeight(section) + bodySectionGap
	}
	slot := rect{x: body.x + bodyInset + fieldsetPadding, y: y + fieldsetPadding + fieldsetTitleH + fieldsetDescH*2, w: 190, h: touchTargetH}
	return slot, true
}

func updateButtonRect(width int, height int) rect {
	slot, _ := nativeControlSlotRect(nativeSlotUpdate, width, height)
	return slot
}

func ScaleToRenderedLayout(viewWidth int, viewHeight int, gameWidth int, gameHeight int) (float64, float64) {
	if gameWidth <= 0 || gameHeight <= 0 {
		return 1, 1
	}
	return float64(viewWidth) / float64(gameWidth), float64(viewHeight) / float64(gameHeight)
}

func UpdateButtonX(width int, height int) int {
	return updateButtonRect(width, height).x
}

func UpdateButtonY(width int, height int) int {
	return updateButtonRect(width, height).y
}

func UpdateButtonW(width int, height int) int {
	return updateButtonRect(width, height).w
}

func UpdateButtonH(width int, height int) int {
	return updateButtonRect(width, height).h
}

func UpdateButtonViewX(viewWidth int, viewHeight int, gameWidth int, gameHeight int) int {
	scaleX, _ := ScaleToRenderedLayout(viewWidth, viewHeight, gameWidth, gameHeight)
	return int(float64(UpdateButtonX(gameWidth, gameHeight)) * scaleX)
}

func UpdateButtonViewY(viewWidth int, viewHeight int, gameWidth int, gameHeight int) int {
	_, scaleY := ScaleToRenderedLayout(viewWidth, viewHeight, gameWidth, gameHeight)
	return int(float64(UpdateButtonY(gameWidth, gameHeight)) * scaleY)
}

func UpdateButtonViewW(viewWidth int, viewHeight int, gameWidth int, gameHeight int) int {
	scaleX, _ := ScaleToRenderedLayout(viewWidth, viewHeight, gameWidth, gameHeight)
	return int(float64(UpdateButtonW(gameWidth, gameHeight)) * scaleX)
}

func UpdateButtonViewH(viewWidth int, viewHeight int, gameWidth int, gameHeight int) int {
	_, scaleY := ScaleToRenderedLayout(viewWidth, viewHeight, gameWidth, gameHeight)
	return int(float64(UpdateButtonH(gameWidth, gameHeight)) * scaleY)
}
