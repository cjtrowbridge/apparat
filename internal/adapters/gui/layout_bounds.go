//go:build gui

// Computes bounded tab and content geometry.
// Dependencies: github.com/ebitenui/ebitenui/input, github.com/ebitenui/ebitenui/widget, github.com/hajimehoshi/ebiten/v2.
package gui

import (
	"github.com/cjtrowbridge/apparat/internal/hud"
	"image"
	"image/color"
	"strings"
	"unicode/utf8"

	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	ebitentext "github.com/hajimehoshi/ebiten/v2/text/v2"
)

type boundedPreferredWidget struct {
	child    widget.PreferredSizeLocateableWidget
	maxWidth func() int
}

func boundPreferredWidth(child widget.PreferredSizeLocateableWidget, maxWidth func() int) widget.PreferredSizeLocateableWidget {
	return &boundedPreferredWidget{child: child, maxWidth: maxWidth}
}

func (b *boundedPreferredWidget) GetWidget() *widget.Widget {
	return b.child.GetWidget()
}

func (b *boundedPreferredWidget) PreferredSize() (int, int) {
	width, height := b.child.PreferredSize()
	if b.maxWidth == nil {
		return width, height
	}
	limit := b.maxWidth()
	if limit > 0 && width > limit {
		width = limit
	}
	return width, height
}

func (b *boundedPreferredWidget) SetLocation(rect image.Rectangle) {
	b.child.SetLocation(rect)
}

func (b *boundedPreferredWidget) Validate() {
	b.child.Validate()
}

func (b *boundedPreferredWidget) RequestRelayout() {
	{
		relayoutable, ok := b.child.(widget.Relayoutable)
		if ok {
			relayoutable.RequestRelayout()
		}
	}
}

func (b *boundedPreferredWidget) Render(screen *ebiten.Image) {
	{
		renderer, ok := b.child.(widget.Renderer)
		if ok {
			renderer.Render(screen)
		}
	}
}

func (b *boundedPreferredWidget) Update(updObj *widget.UpdateObject) {
	{
		updater, ok := b.child.(widget.Updater)
		if ok {
			updater.Update(updObj)
		}
	}
}

func (b *boundedPreferredWidget) SetupInputLayer(def input.DeferredSetupInputLayerFunc) {
	{
		layerer, ok := b.child.(input.Layerer)
		if ok {
			layerer.SetupInputLayer(def)
		}
	}
}

func (b *boundedPreferredWidget) GetFocusers() []widget.Focuser {
	{
		container, ok := b.child.(interface{ GetFocusers() []widget.Focuser })
		if ok {
			return container.GetFocusers()
		}
	}
	return nil
}

func (b *boundedPreferredWidget) GetDropTargets() []widget.HasWidget {
	{
		dropper, ok := b.child.(widget.Dropper)
		if ok {
			return dropper.GetDropTargets()
		}
	}
	return nil
}

func (game *Game) hudPreferredWidth() int {
	return max(120, game.contentWidth()-(windowMargin*2))
}

func (game *Game) hudTextMaxWidth() float64 {
	return float64(max(120, game.hudPreferredWidth()-72))
}

func (game *Game) detailTextMaxWidth() float64 {
	// Keep readable detail text outside the floating PTT capture region.
	width := game.hudPreferredWidth() - 72 - (2*pttRadius + pttMargin)
	if !game.collapsed() {
		width -= game.masterListWidth() + 40
	}
	return float64(max(120, width))
}

func (game *Game) hudText(value string) *widget.Text {
	return widget.NewText(
		widget.TextOpts.Text(value, game.theme.ButtonTheme.TextFace, color.White),
		widget.TextOpts.MaxWidth(game.hudTextMaxWidth()),
	)
}

func (game *Game) detailText(value string) *widget.Text {
	return widget.NewText(
		widget.TextOpts.Text(value, game.theme.ButtonTheme.TextFace, color.White),
		widget.TextOpts.MaxWidth(game.detailTextMaxWidth()),
	)
}

// Keep words intact where possible, then break long names at rune boundaries.
func wrapSelectorLabel(label string, face ebitentext.Face, width int) string {
	fits := func(line string) bool {
		measured, _ := ebitentext.Measure(line, face, 0)
		return measured <= float64(width)
	}
	if fits(label) {
		return label
	}

	var lines []string
	line := ""
	for _, word := range strings.Fields(label) {
		if line != "" {
			candidate := line + " " + word
			if fits(candidate) {
				line = candidate
				continue
			}
			lines = append(lines, line)
			line = ""
		}
		if fits(word) {
			line = word
			continue
		}
		for _, letter := range word {
			candidate := line + string(letter)
			if line != "" && !fits(candidate) {
				lines = append(lines, line)
				line = string(letter)
			} else {
				line = candidate
			}
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// Preserve every character, including repeated spaces in directory names.
func wrapSelectorPath(path string, face ebitentext.Face, width int) string {
	var result strings.Builder
	start := 0
	for offset := range path {
		_, bytes := utf8.DecodeRuneInString(path[offset:])
		end := offset + bytes
		size, _ := ebitentext.Measure(path[start:end], face, 0)
		if offset > start && size > float64(width) {
			result.WriteString(path[start:offset])
			result.WriteByte('\n')
			start = offset
		}
	}
	result.WriteString(path[start:])
	return result.String()
}

const railGutter = tabIconButtonWidth + bodyGap

func placementLabel(setting string) string {
	switch setting {
	case "bottom":
		return "Bottom"
	case "left_rail":
		return "Left"
	case "right_rail":
		return "Right"
	case "responsive":
		return "Responsive"
	default:
		return "Top"
	}
}

func (game *Game) bottomTabs() bool {
	return game.prefs.Placement == "bottom" || game.prefs.Placement == "responsive" && game.height > game.width
}

// Fall back to the scrollable top strip when a rail cannot fit every tab.
func (game *Game) sideRail() string {
	mode := game.prefs.Placement
	if game.bottomTabs() || mode == "top" || mode == "" {
		return ""
	}
	count := len(game.shell.Snapshot().Tabs)
	if game.height-tabTop-diagnosticsHeight < count*tabHeight+(count-1)*tabGap {
		return ""
	}
	if mode == "responsive" {
		if game.width-railGutter < collapsedWidth {
			return ""
		}
		return "left_rail"
	}
	return mode
}
func (game *Game) contentWidth() int {
	if game.sideRail() != "" {
		return game.width - railGutter
	}
	return game.width
}

func (game *Game) pttCenter() (int, int) {
	if game.actions.mode != "" {
		return game.width - windowMargin - pttRadius, game.height - pttRadius - 4
	}
	x := game.width - pttMargin - pttRadius
	if game.sideRail() == "right_rail" {
		x -= railGutter
	}
	y := game.height - diagnosticsHeight - pttMargin - pttRadius
	if game.bottomTabs() {
		y -= tabHeight + bodyGap
	}
	return x, y
}

func (game *Game) detailSections(tabData hud.Tab) []hud.Section {
	if len(tabData.Sections) == 0 {
		return nil
	}
	index := game.selectedSectionIndex(tabData)
	if index < 0 {
		return nil
	}
	return []hud.Section{tabData.Sections[index]}
}

func (game *Game) collapsed() bool {
	return game.contentWidth() < collapsedWidth
}
