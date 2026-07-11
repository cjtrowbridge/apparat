//go:build gui

package gui

import (
	"image"
	"image/color"
)

func appIconImages() []image.Image {
	return []image.Image{
		appIconImage(32),
		appIconImage(64),
	}
}

func appIconImage(size int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	blue := color.RGBA{B: 255, A: 255}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.SetRGBA(x, y, white)
		}
	}
	fillCircle(img, size/2, size/2, size*29/100, blue)
	toothW := max(4, size/8)
	toothL := max(6, size/5)
	fillRect(img, size/2-toothW/2, size/10, toothW, toothL, blue)
	fillRect(img, size/2-toothW/2, size-size/10-toothL, toothW, toothL, blue)
	fillRect(img, size/10, size/2-toothW/2, toothL, toothW, blue)
	fillRect(img, size-size/10-toothL, size/2-toothW/2, toothL, toothW, blue)
	cut := size / 5
	fillCircle(img, size/2, size/2, cut, white)
	return img
}

func fillCircle(img *image.RGBA, cx int, cy int, radius int, c color.RGBA) {
	r2 := radius * radius
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			if (image.Point{X: x, Y: y}).In(img.Bounds()) && (x-cx)*(x-cx)+(y-cy)*(y-cy) <= r2 {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func fillRect(img *image.RGBA, x int, y int, width int, height int, c color.RGBA) {
	for yy := y; yy < y+height; yy++ {
		for xx := x; xx < x+width; xx++ {
			if (image.Point{X: xx, Y: yy}).In(img.Bounds()) {
				img.SetRGBA(xx, yy, c)
			}
		}
	}
}
