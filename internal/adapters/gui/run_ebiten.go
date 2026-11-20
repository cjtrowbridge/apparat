//go:build gui

// Starts the native Ebitengine GUI loop.
// Dependencies: github.com/hajimehoshi/ebiten/v2.
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
	game := NewGameWithRuntimeInfo(info)
	if gamescopeSession() {
		fillGamescope()
	} else {
		game.restoreWindow()
	}
	game.initDesktopUpdates(ctx)
	game.runContext = ctx
	defer game.Close()
	err := ebiten.RunGame(game)
	if err == nil && game.desktopUpdates != nil && game.desktopUpdates.restart {
		return ErrRestart
	}
	return err
}
