//go:build gui

package gui

import (
	"image/color"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"golang.org/x/image/font/basicfont"
)

func createUITheme() *widget.Theme {
	theme := widget.Theme{}

	// Basic font setup
	face := basicfont.Face7x13
	
	// Define colors from existing definitions
	bgColor := color.RGBA{R: 15, G: 18, B: 28, A: 255}
	panelBgColor := color.RGBA{R: 25, G: 30, B: 46, A: 255}
	accentColor := color.RGBA{R: 62, G: 96, B: 150, A: 255}
	textColor := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	disabledTextColor := color.RGBA{R: 148, G: 160, B: 184, A: 255}
	borderColor := color.RGBA{R: 40, G: 45, B: 60, A: 255}
	
	// Button styles
	buttonImage := &widget.ButtonImage{
		Idle:     image.NewNineSliceColor(panelBgColor),
		Hover:    image.NewNineSliceColor(accentColor),
		Pressed:  image.NewNineSliceColor(accentColor),
		Disabled: image.NewNineSliceColor(bgColor),
	}

	buttonTextColor := &widget.ButtonTextColor{
		Idle:     textColor,
		Hover:    textColor,
		Pressed:  textColor,
		Disabled: disabledTextColor,
	}

	theme.Button = &widget.ButtonTheme{
		Image:   buttonImage,
		Text:    buttonTextColor,
		Face:    face,
		Padding: widget.Insets{Left: 16, Right: 16, Top: 8, Bottom: 8},
	}

	// TabBook styles
	theme.TabBook = &widget.TabBookTheme{
		TabButtonImage: &widget.ButtonImage{
			Idle:     image.NewNineSliceColor(bgColor),
			Hover:    image.NewNineSliceColor(panelBgColor),
			Pressed:  image.NewNineSliceColor(accentColor),
			Disabled: image.NewNineSliceColor(bgColor),
		},
		TabButtonText: &widget.ButtonTextColor{
			Idle:     textColor,
			Hover:    textColor,
			Pressed:  textColor,
			Disabled: disabledTextColor,
		},
		TabButtonSpacing: 4,
		TabButtonPadding: widget.Insets{Left: 12, Right: 12, Top: 8, Bottom: 8},
		TabButtonFace:    face,
		TabPanelPadding:  widget.Insets{Left: 0, Right: 0, Top: 8, Bottom: 0},
	}

	// Panel styles
	theme.Panel = &widget.PanelTheme{
		Image:   image.NewNineSliceColor(panelBgColor),
		Padding: widget.Insets{Left: 12, Right: 12, Top: 12, Bottom: 12},
	}

	// Scroll Container
	theme.ScrollContainer = &widget.ScrollContainerTheme{}
	
	theme.Slider = &widget.SliderTheme{
		Track: &widget.SliderTrackImage{
			Idle:     image.NewNineSliceColor(borderColor),
			Hover:    image.NewNineSliceColor(borderColor),
		},
		Handle: &widget.ButtonImage{
			Idle:     image.NewNineSliceColor(accentColor),
			Hover:    image.NewNineSliceColor(textColor),
			Pressed:  image.NewNineSliceColor(textColor),
			Disabled: image.NewNineSliceColor(disabledTextColor),
		},
		HandleSize: 6,
		TrackPadding: widget.Insets{Left: 2, Right: 2, Top: 2, Bottom: 2},
	}

	// Checkbox styles
	theme.Checkbox = &widget.CheckboxTheme{
		Image: &widget.ButtonImage{
			Idle:     image.NewNineSliceColor(panelBgColor),
			Hover:    image.NewNineSliceColor(accentColor),
			Pressed:  image.NewNineSliceColor(accentColor),
			Disabled: image.NewNineSliceColor(bgColor),
		},
		Graphic: &widget.CheckboxGraphicImage{
			Checked:   image.NewNineSliceColor(textColor),
			Unchecked: image.NewNineSliceColor(color.Transparent),
		},
		Spacing: 8,
		Padding: widget.Insets{Left: 8, Right: 8, Top: 8, Bottom: 8},
		Face:    face,
		TextColor: &widget.CheckboxTextColor{
			Idle:     textColor,
			Hover:    textColor,
			Disabled: disabledTextColor,
		},
	}

	// Window styles
	theme.Window = &widget.WindowTheme{
		TitleBar: &widget.WindowTitleBarTheme{
			Image:   image.NewNineSliceColor(accentColor),
			Padding: widget.Insets{Left: 8, Right: 8, Top: 4, Bottom: 4},
			TitleColor: textColor,
			FontFace: face,
		},
		Background: image.NewNineSliceColor(bgColor),
	}

	return &theme
}
