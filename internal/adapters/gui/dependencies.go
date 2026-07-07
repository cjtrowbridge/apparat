//go:build gui

package gui

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

var (
	_ = ebiten.DeviceScaleFactor
	_ = widget.NewContainer
)
