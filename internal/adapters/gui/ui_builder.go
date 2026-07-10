//go:build gui

package gui

import (
	"fmt"
	img "image"
	"image/color"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
)

// rebuildUI constructs the complete EbitenUI widget tree from the current HUD snapshot.
// In the future, this should intelligently update existing widgets instead of rebuilding
// from scratch, but for this initial migration, rebuilding ensures layout correctness.
func (game *Game) rebuildUI(snapshot hud.Snapshot) {
	// Create the root container with a vertical layout
	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(tabGap),
		)),
	)

	// Build the top tab bar
	tabs := game.buildTabs(snapshot)
	tabBook := widget.NewTabBook(
		widget.TabBookOpts.Tabs(tabs...),
		widget.TabBookOpts.TabButtonSpacing(tabGap),
		widget.TabBookOpts.TabButtonMinSize(&img.Point{X: 0, Y: tabHeight}),
		widget.TabBookOpts.TabSelectedHandler(func(args *widget.TabBookTabSelectedEventArgs) {
			for i, t := range tabs {
				if t == args.Tab {
					_ = game.shell.SelectTab(i)
					break
				}
			}
		}),
	)

	root.AddChild(tabBook)

	// Set the tab book to the active index
	if len(tabs) > 0 && snapshot.ActiveIndex >= 0 && snapshot.ActiveIndex < len(tabs) {
		tabBook.SetTab(tabs[snapshot.ActiveIndex])
	}

	// Update the game UI instance
	game.ui = &ebitenui.UI{
		Container:    root,
		PrimaryTheme: game.theme,
	}
}

func (game *Game) buildTabs(snapshot hud.Snapshot) []*widget.TabBookTab {
	tabs := make([]*widget.TabBookTab, 0, len(snapshot.Tabs))
	for _, tabData := range snapshot.Tabs {
		label := fmt.Sprintf("%s %s", tabData.Descriptor.Glyph, tabData.Title())

		tab := widget.NewTabBookTab(
			widget.TabBookTabOpts.Label(label),
			widget.TabBookTabOpts.ContainerOpts(
				widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
			),
		)

		if tabData.Descriptor.ID == hud.TabSettings {
			tab.AddChild(game.buildSettingsTab(tabData))
		} else {
			// Placeholder for other tabs (master-detail layouts will go here)
			placeholderText := widget.NewText(
				widget.TextOpts.Text(fmt.Sprintf("%s Panel (EbitenUI)", tabData.Title()), game.theme.ButtonTheme.TextFace, color.White),
			)
			tab.AddChild(placeholderText)
		}

		tabs = append(tabs, tab)
	}
	return tabs
}

func (game *Game) buildSettingsTab(tabData hud.Tab) widget.PreferredSizeLocateableWidget {
	content := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(bodyGap),
		)),
	)

	scrollContainer := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(content),
		widget.ScrollContainerOpts.Image(createScrollContainerImage()),
	)

	// Example: Add an Update Check section manually for now
	updateSection := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(game.theme.PanelTheme.BackgroundImage),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(&widget.Insets{Left: 12, Right: 12, Top: 12, Bottom: 12}),
			widget.RowLayoutOpts.Spacing(4),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)

	title := widget.NewText(
		widget.TextOpts.Text("Updates", game.theme.ButtonTheme.TextFace, color.White),
	)
	updateSection.AddChild(title)

	updateBtn := widget.NewButton(
		widget.ButtonOpts.Text("Check for update", game.theme.ButtonTheme.TextFace, game.theme.ButtonTheme.TextColor),
		widget.ButtonOpts.Image(game.theme.ButtonTheme.Image),
		widget.ButtonOpts.TextPadding(game.theme.ButtonTheme.TextPadding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			if game.onCheckForUpdate != nil {
				game.onCheckForUpdate()
			}
		}),
	)
	updateSection.AddChild(updateBtn)

	content.AddChild(updateSection)
	return scrollContainer
}
