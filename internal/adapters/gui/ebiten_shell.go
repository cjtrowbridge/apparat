//go:build gui

package gui

import (
	"image/color"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	shell  hud.Shell
	width  int
	height int
}

func NewGame() *Game {
	return &Game{
		shell:  hud.NewShell(),
		width:  1280,
		height: 800,
	}
}

func (game *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyControl) && ebiten.IsKeyPressed(ebiten.KeyPageDown) {
		game.shell.NextTab()
	}
	if ebiten.IsKeyPressed(ebiten.KeyControl) && ebiten.IsKeyPressed(ebiten.KeyPageUp) {
		game.shell.PreviousTab()
	}
	return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 15, G: 18, B: 28, A: 255})
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
