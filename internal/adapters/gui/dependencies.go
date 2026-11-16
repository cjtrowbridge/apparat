//go:build gui

// Records the package dependency boundary.
// Dependencies: github.com/ebitenui/ebitenui/widget, github.com/hajimehoshi/ebiten/v2.
package gui

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

var (
	_ = ebiten.DeviceScaleFactor
	_ = widget.NewContainer
)
