//go:build gui

// Selects the compositor's native window geometry in Gaming Mode.
// Ordinary desktop sessions retain their saved placement and resizable windows.
package gui

import (
	"os"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
)

func gamescopeSession() bool {
	return runtime.GOOS == "linux" && os.Getenv("GAMESCOPE_WAYLAND_DISPLAY") != ""
}

func fillGamescope() {
	monitor := ebiten.Monitor()
	if monitor != nil {
		width, height := monitor.Size()
		if width > 0 && height > 0 {
			ebiten.SetWindowSize(width, height)
		}
	}
	ebiten.SetFullscreen(true)
}
