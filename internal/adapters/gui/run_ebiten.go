//go:build gui

package gui

import (
	"context"

	"github.com/hajimehoshi/ebiten/v2"
)

func Run(ctx context.Context) error {
	return RunWithRuntimeInfo(ctx, defaultRuntimeInfo())
}

func RunWithRuntimeInfo(ctx context.Context, info RuntimeInfo) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	ebiten.SetWindowTitle("Apparat")
	ebiten.SetWindowIcon(appIconImages())
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return ebiten.RunGame(NewGameWithRuntimeInfo(info))
}
