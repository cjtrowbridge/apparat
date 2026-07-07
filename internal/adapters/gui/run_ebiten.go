//go:build gui

package gui

import (
	"context"

	"github.com/hajimehoshi/ebiten/v2"
)

func Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	ebiten.SetWindowTitle("Apparat")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return ebiten.RunGame(NewGame())
}
