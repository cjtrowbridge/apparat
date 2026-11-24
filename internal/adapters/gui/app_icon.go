//go:build gui

// Loads the desktop window icon from the rasterized Apparat gear.
// The embedded PNG preserves the transparent hub and outlined teeth.
package gui

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
)

//go:embed logo.png
var appIconPNG []byte

func appIconImages() []image.Image {
	icon, err := png.Decode(bytes.NewReader(appIconPNG))
	if err != nil {
		panic("invalid embedded Apparat icon: " + err.Error())
	}
	return []image.Image{icon}
}
