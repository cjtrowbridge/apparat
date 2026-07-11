//go:build gui

package gui

import (
	"bytes"
	_ "embed"
	"image/color"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	ebitentext "github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/basicfont"
)

//go:embed DejaVuSans.ttf
var dejavuSansTTF []byte

func createUITheme() *widget.Theme {
	theme := widget.Theme{}

	// TTF Font Setup using embedded DejaVu Sans for full Unicode support
	var face ebitentext.Face
	source, err := ebitentext.NewGoTextFaceSource(bytes.NewReader(dejavuSansTTF))
	if err == nil {
		face = &ebitentext.GoTextFace{
			Source: source,
			Size:   14,
		}
	} else {
		// Fallback to basic font if loading fails
		face = ebitentext.NewGoXFace(basicfont.Face7x13)
	}
	facePtr := &face

	// Define colors from existing definitions
	bgColor := color.RGBA{R: 15, G: 18, B: 28, A: 255}
	panelBgColor := color.RGBA{R: 25, G: 30, B: 46, A: 255}
	accentColor := color.RGBA{R: 62, G: 96, B: 150, A: 255}
	textColor := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	disabledTextColor := color.RGBA{R: 148, G: 160, B: 184, A: 255}
	borderColor := color.RGBA{R: 40, G: 45, B: 60, A: 255}

	// Button styles
	buttonBgColor := color.RGBA{R: 45, G: 55, B: 78, A: 255}
	theme.ButtonTheme = &widget.ButtonParams{
		Image: &widget.ButtonImage{
			Idle:     image.NewNineSliceColor(buttonBgColor),
			Hover:    image.NewNineSliceColor(accentColor),
			Pressed:  image.NewNineSliceColor(accentColor),
			Disabled: image.NewNineSliceColor(bgColor),
		},
		TextColor: &widget.ButtonTextColor{
			Idle:     textColor,
			Hover:    textColor,
			Pressed:  textColor,
			Disabled: disabledTextColor,
		},
		TextFace:    facePtr,
		TextPadding: &widget.Insets{Left: 16, Right: 16, Top: 8, Bottom: 8},
	}

	// TabBook styles
	tabSpacing := 4
	contentSpacing := 8
	theme.TabbookTheme = &widget.TabBookParams{
		TabButton: &widget.ButtonParams{
			Image: &widget.ButtonImage{
				Idle:     image.NewNineSliceColor(bgColor),
				Hover:    image.NewNineSliceColor(panelBgColor),
				Pressed:  image.NewNineSliceColor(accentColor),
				Disabled: image.NewNineSliceColor(bgColor),
			},
			TextColor: &widget.ButtonTextColor{
				Idle:     textColor,
				Hover:    textColor,
				Pressed:  textColor,
				Disabled: disabledTextColor,
			},
			TextFace:    facePtr,
			TextPadding: &widget.Insets{Left: 12, Right: 12, Top: 8, Bottom: 8},
		},
		TabSpacing:     &tabSpacing,
		ContentSpacing: &contentSpacing,
		ContentPadding: &widget.Insets{Left: 0, Right: 0, Top: 8, Bottom: 0},
	}

	// Panel styles
	theme.PanelTheme = &widget.PanelParams{
		BackgroundImage: image.NewNineSliceColor(panelBgColor),
	}
	theme.LabelTheme = &widget.LabelParams{
		Color: &widget.LabelColor{
			Idle:     textColor,
			Disabled: disabledTextColor,
		},
		Face:    facePtr,
		Padding: &widget.Insets{Left: 8, Right: 8, Top: 8, Bottom: 8},
	}

	theme.SliderTheme = &widget.SliderParams{
		TrackImage: &widget.SliderTrackImage{
			Idle:  image.NewNineSliceColor(borderColor),
			Hover: image.NewNineSliceColor(borderColor),
		},
		HandleImage: &widget.ButtonImage{
			Idle:     image.NewNineSliceColor(accentColor),
			Hover:    image.NewNineSliceColor(textColor),
			Pressed:  image.NewNineSliceColor(textColor),
			Disabled: image.NewNineSliceColor(disabledTextColor),
		},
		TrackPadding: &widget.Insets{Left: 2, Right: 2, Top: 2, Bottom: 2},
	}

	theme.CheckboxTheme = &widget.CheckboxParams{
		Image: &widget.CheckboxImage{
			Unchecked: image.NewNineSliceColor(panelBgColor),
			Checked:   image.NewNineSliceColor(accentColor),
		},
		Label: &widget.LabelParams{
			Color: &widget.LabelColor{
				Idle:     textColor,
				Disabled: disabledTextColor,
			},
			Face:    facePtr,
			Padding: &widget.Insets{Left: 8, Right: 8, Top: 8, Bottom: 8},
		},
	}

	return &theme
}

func createScrollContainerImage() *widget.ScrollContainerImage {
	return &widget.ScrollContainerImage{
		Idle: image.NewNineSliceColor(color.RGBA{0, 0, 0, 0}),
		Mask: image.NewNineSliceColor(color.RGBA{255, 255, 255, 255}),
	}
}
