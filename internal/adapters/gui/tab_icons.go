//go:build gui

// Loads the fixed tab artwork once and supplies icon-only button contents.
// Uses embedded PNGs, Ebitengine images and the existing EbitenUI button API.
package gui

import (
	"bytes"
	"embed"
	"fmt"
	"image/png"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed emoji/*.png
var tabArtwork embed.FS

// Owned by the GUI thread; retained across widget rebuilds for process lifetime.
var tabIcons = map[string]*ebiten.Image{}

const tabIconSize = 28
const tabIconButtonWidth = 54

func tabTargetWidth(available, count int) int {
	if count <= 5 {
		return tabIconButtonWidth
	}
	return max(44, min(tabIconButtonWidth, (available-tabGap*(count-1))/max(1, count)))
}

func (game *Game) tabIconContent(tab hud.TabDescriptor) widget.ButtonOpt {
	if tab.ID == hud.TabComrades {
		return widget.ButtonOpts.Text(tab.Glyph, createTextFace(tabIconSize), game.theme.ButtonTheme.TextColor)
	}
	names := map[hud.TabID]string{
		"stmpck": "1f489", hud.TabProjects: "1f528", hud.TabCluster: "1f5a7",
		hud.TabResearch: "1f9ea", hud.TabSettings: "2699",
	}
	return widget.ButtonOpts.Graphic(&widget.GraphicImage{Idle: emojiIcon(names[tab.ID])})
}

func emojiIcon(name string) *ebiten.Image {
	icon := tabIcons[name]
	if icon == nil {
		data, err := tabArtwork.ReadFile("emoji/" + name + ".png")
		if err != nil {
			panic(fmt.Errorf("load tab icon %s: %w", name, err))
		}
		decoded, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			panic(fmt.Errorf("decode tab icon %s: %w", name, err))
		}
		source := ebiten.NewImageFromImage(decoded)
		icon = ebiten.NewImage(tabIconSize, tabIconSize)
		options := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
		options.GeoM.Scale(float64(tabIconSize)/float64(source.Bounds().Dx()), float64(tabIconSize)/float64(source.Bounds().Dy()))
		icon.DrawImage(source, options)
		tabIcons[name] = icon
	}
	return icon
}
