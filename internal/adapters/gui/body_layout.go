//go:build gui

package gui

import (
	"fmt"
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
)

type rect struct {
	x int
	y int
	w int
	h int
}

func drawActiveTab(screen *ebiten.Image, snapshot hud.Snapshot, width int, height int) {
	body := tabBodyRect(width, height)
	ebitenutil.DrawRect(screen, float64(body.x), float64(body.y), float64(body.w), float64(body.h), panelColor)
	tab := snapshot.ActiveTab()
	if tab.ID() == hud.TabSettings {
		drawSettingsBody(screen, tab, body)
		return
	}
	drawMasterDetailBody(screen, tab, body)
}

func tabBodyRect(width int, height int) rect {
	y := tabTop + tabHeight + bodyGap
	return rect{x: windowMargin, y: y, w: width - windowMargin*2, h: height - y - diagnosticsHeight}
}

func drawSettingsBody(screen *ebiten.Image, tab hud.Tab, body rect) {
	x := body.x + bodyInset
	y := body.y + bodyInset
	w := body.w - bodyInset*2
	ebitenutil.DebugPrintAt(screen, tab.Title(), x, y)
	ebitenutil.DebugPrintAt(screen, truncateText(tab.Summary, w), x, y+24)
	y += 58
	for _, section := range tab.Sections {
		h := fieldsetHeight(section)
		drawFieldset(screen, rect{x: x, y: y, w: w, h: h}, section)
		y += h + bodySectionGap
		if y > body.y+body.h-fieldsetMinH {
			break
		}
	}
}

func drawMasterDetailBody(screen *ebiten.Image, tab hud.Tab, body rect) {
	list, detail := masterDetailRects(body)
	drawListPane(screen, tab, list)
	drawDetailPane(screen, tab, detail)
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

func drawListPane(screen *ebiten.Image, tab hud.Tab, pane rect) {
	ebitenutil.DrawRect(screen, float64(pane.x), float64(pane.y), float64(pane.w), float64(pane.h), listPaneColor)
	drawBorder(screen, pane)
	ebitenutil.DebugPrintAt(screen, tab.Title(), pane.x+fieldsetPadding, pane.y+fieldsetPadding)
	y := pane.y + fieldsetPadding + 28
	items := listItemsForTab(tab)
	for index, item := range items {
		if index == 0 {
			ebitenutil.DrawRect(screen, float64(pane.x+6), float64(y-6), float64(pane.w-12), float64(touchTargetH), selectedItemColor)
		}
		ebitenutil.DebugPrintAt(screen, truncateText(item, pane.w-fieldsetPadding*2), pane.x+fieldsetPadding, y)
		y += touchTargetH
		if y > pane.y+pane.h-touchTargetH {
			break
		}
	}
}

func drawDetailPane(screen *ebiten.Image, tab hud.Tab, pane rect) {
	ebitenutil.DrawRect(screen, float64(pane.x), float64(pane.y), float64(pane.w), float64(pane.h), fieldsetColor)
	drawBorder(screen, pane)
	x := pane.x + fieldsetPadding
	y := pane.y + fieldsetPadding
	ebitenutil.DebugPrintAt(screen, "Placeholder Detail", x, y)
	ebitenutil.DebugPrintAt(screen, truncateText(tab.Summary, pane.w-fieldsetPadding*2), x, y+24)
	y += 58
	for _, section := range tab.Sections {
		h := fieldsetHeight(section)
		drawFieldset(screen, rect{x: x, y: y, w: pane.w - fieldsetPadding*2, h: h}, section)
		y += h + bodySectionGap
		if y > pane.y+pane.h-fieldsetMinH {
			break
		}
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
		ebitenutil.DebugPrintAt(screen, truncateText(section.Description, area.w-fieldsetPadding*2), x, y)
		y += fieldsetDescH
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
		h += fieldsetDescH
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

func drawBorder(screen *ebiten.Image, area rect) {
	ebitenutil.DrawLine(screen, float64(area.x), float64(area.y), float64(area.x+area.w), float64(area.y), bodyBorderColor)
	ebitenutil.DrawLine(screen, float64(area.x), float64(area.y+area.h), float64(area.x+area.w), float64(area.y+area.h), bodyBorderColor)
	ebitenutil.DrawLine(screen, float64(area.x), float64(area.y), float64(area.x), float64(area.y+area.h), bodyBorderColor)
	ebitenutil.DrawLine(screen, float64(area.x+area.w), float64(area.y), float64(area.x+area.w), float64(area.y+area.h), bodyBorderColor)
}

func updateButtonRect(width int, height int) rect {
	body := tabBodyRect(width, height)
	y := body.y + bodyInset + 58
	sections := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[6].Sections
	for _, section := range sections[:len(sections)-1] {
		y += fieldsetHeight(section) + bodySectionGap
	}
	return rect{x: body.x + bodyInset + fieldsetPadding, y: y + fieldsetPadding + fieldsetTitleH + fieldsetDescH, w: 190, h: touchTargetH}
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
