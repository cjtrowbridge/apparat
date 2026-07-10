//go:build gui

package gui

import (
	"image/color"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	ebitentext "github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/basicfont"
)

func createBorderedNineSlice(fill color.Color, border color.Color) *image.NineSlice {
	i := ebiten.NewImage(3, 3)
	i.Fill(border)
	i.Set(1, 1, fill)
	return image.NewNineSlice(i, [3]int{1, 1, 1}, [3]int{1, 1, 1})
}

func createUITheme() *widget.Theme {
	theme := widget.Theme{}

	// Basic font setup
	var f ebitentext.Face = ebitentext.NewGoXFace(basicfont.Face7x13)
	face := &f

	// Define colors from existing definitions
	bgColor := color.RGBA{R: 15, G: 18, B: 28, A: 255}
	panelBgColor := color.RGBA{R: 25, G: 30, B: 46, A: 255}
	accentColor := color.RGBA{R: 62, G: 96, B: 150, A: 255}
	textColor := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	disabledTextColor := color.RGBA{R: 148, G: 160, B: 184, A: 255}
	borderColor := color.RGBA{R: 40, G: 45, B: 60, A: 255}

	// Button styles
	theme.ButtonTheme = &widget.ButtonParams{
		Image: &widget.ButtonImage{
			Idle:     createBorderedNineSlice(panelBgColor, borderColor),
			Hover:    createBorderedNineSlice(accentColor, borderColor),
			Pressed:  createBorderedNineSlice(accentColor, borderColor),
			Disabled: createBorderedNineSlice(bgColor, borderColor),
		},
		TextColor: &widget.ButtonTextColor{
			Idle:     textColor,
			Hover:    textColor,
			Pressed:  textColor,
			Disabled: disabledTextColor,
		},
		TextFace:    face,
		TextPadding: &widget.Insets{Left: 16, Right: 16, Top: 8, Bottom: 8},
	}

	// TabBook styles
	tabSpacing := 4
	contentSpacing := 8
	theme.TabbookTheme = &widget.TabBookParams{
		TabButton: &widget.ButtonParams{
			Image: &widget.ButtonImage{
				Idle:     createBorderedNineSlice(bgColor, borderColor),
				Hover:    createBorderedNineSlice(panelBgColor, borderColor),
				Pressed:  createBorderedNineSlice(accentColor, borderColor),
				Disabled: createBorderedNineSlice(bgColor, borderColor),
			},
			TextColor: &widget.ButtonTextColor{
				Idle:     textColor,
				Hover:    textColor,
				Pressed:  textColor,
				Disabled: disabledTextColor,
			},
			TextFace:    face,
			TextPadding: &widget.Insets{Left: 12, Right: 12, Top: 8, Bottom: 8},
		},
		TabSpacing:     &tabSpacing,
		ContentSpacing: &contentSpacing,
		ContentPadding: &widget.Insets{Left: 0, Right: 0, Top: 8, Bottom: 0},
	}

	// Panel styles
	theme.PanelTheme = &widget.PanelParams{
		BackgroundImage: createBorderedNineSlice(panelBgColor, borderColor),
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
			Face:    face,
			Padding: &widget.Insets{Left: 8, Right: 8, Top: 8, Bottom: 8},
		},
	}

	return &theme
}

func createScrollContainerImage() *widget.ScrollContainerImage {
	borderColor := color.RGBA{R: 40, G: 45, B: 60, A: 255}
	return &widget.ScrollContainerImage{
		Idle: createBorderedNineSlice(color.RGBA{0, 0, 0, 0}, borderColor),
		Mask: createBorderedNineSlice(color.RGBA{0, 0, 0, 0}, borderColor),
	}
}
