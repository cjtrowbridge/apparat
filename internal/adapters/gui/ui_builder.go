//go:build gui

package gui

import (
	"fmt"
	img "image"
	"image/color"
	"strings"

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
			tab.AddChild(game.buildMasterDetailTab(tabData))
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
			widget.RowLayoutOpts.Padding(&widget.Insets{Left: 12, Right: 12, Top: 12, Bottom: 12}),
		)),
	)

	if tabData.Summary != "" {
		summary := widget.NewText(
			widget.TextOpts.Text(tabData.Summary, game.theme.ButtonTheme.TextFace, color.White),
		)
		content.AddChild(summary)
	}

	for _, section := range tabData.Sections {
		sectionContainer := widget.NewContainer(
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
			widget.TextOpts.Text(strings.ToUpper(section.Title), game.theme.ButtonTheme.TextFace, color.White),
		)
		sectionContainer.AddChild(title)

		if section.Description != "" {
			desc := widget.NewText(
				widget.TextOpts.Text(section.Description, game.theme.ButtonTheme.TextFace, color.White),
			)
			sectionContainer.AddChild(desc)
		}

		for _, row := range section.Rows {
			rowText := row.Label
			if row.Detail != "" {
				rowText = fmt.Sprintf("%s: %s", row.Label, row.Detail)
			}
			label := widget.NewText(
				widget.TextOpts.Text(rowText, game.theme.ButtonTheme.TextFace, color.White),
			)
			sectionContainer.AddChild(label)
		}

		if strings.ToLower(section.Title) == "updates" {
			updateBtn := widget.NewButton(
				widget.ButtonOpts.Text("Check for update", game.theme.ButtonTheme.TextFace, game.theme.ButtonTheme.TextColor),
				widget.ButtonOpts.Image(game.theme.ButtonTheme.Image),
				widget.ButtonOpts.TextPadding(game.theme.ButtonTheme.TextPadding),
				widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(0, 44)),
				widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
					if game.onCheckForUpdate != nil {
						game.onCheckForUpdate()
					} else {
						args.Button.Text().Label = "Update check simulated"
					}
				}),
			)
			sectionContainer.AddChild(updateBtn)
		}
		content.AddChild(sectionContainer)
	}

	scrollContainer := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(content),
		widget.ScrollContainerOpts.Image(createScrollContainerImage()),
		widget.ScrollContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{StretchHorizontal: true, StretchVertical: true})),
	)

	return scrollContainer
}

func (game *Game) buildMasterDetailTab(tabData hud.Tab) widget.PreferredSizeLocateableWidget {
	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, true}, []bool{true}),
			widget.GridLayoutOpts.Spacing(8, 0),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{StretchHorizontal: true, StretchVertical: true})),
	)

	// Left List Pane
	listContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(4),
			widget.RowLayoutOpts.Padding(&widget.Insets{Left: 8, Right: 8, Top: 8, Bottom: 8}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(170, 0),
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				MaxHeight: 0,
			}),
		),
	)

	for _, section := range tabData.Sections {
		btn := widget.NewButton(
			widget.ButtonOpts.Text(section.Title, game.theme.ButtonTheme.TextFace, game.theme.ButtonTheme.TextColor),
			widget.ButtonOpts.Image(game.theme.ButtonTheme.Image),
			widget.ButtonOpts.TextPadding(game.theme.ButtonTheme.TextPadding),
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.MinSize(0, 44),
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
			),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
				// Detail update logic placeholder
			}),
		)
		listContainer.AddChild(btn)
	}

	listScroll := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(listContainer),
		widget.ScrollContainerOpts.Image(createScrollContainerImage()),
		widget.ScrollContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.GridLayoutData{
			MaxHeight: 0,
		})),
	)
	root.AddChild(listScroll)

	// Right Detail Pane
	detailContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(bodyGap),
			widget.RowLayoutOpts.Padding(&widget.Insets{Left: 12, Right: 12, Top: 12, Bottom: 12}),
		)),
	)

	if tabData.Summary != "" {
		summary := widget.NewText(
			widget.TextOpts.Text(tabData.Summary, game.theme.ButtonTheme.TextFace, color.White),
		)
		detailContainer.AddChild(summary)
	}

	for _, section := range tabData.Sections {
		sectionContainer := widget.NewContainer(
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
			widget.TextOpts.Text(strings.ToUpper(section.Title), game.theme.ButtonTheme.TextFace, color.White),
		)
		sectionContainer.AddChild(title)

		if section.Description != "" {
			desc := widget.NewText(
				widget.TextOpts.Text(section.Description, game.theme.ButtonTheme.TextFace, color.White),
			)
			sectionContainer.AddChild(desc)
		}

		for _, row := range section.Rows {
			rowText := row.Label
			if row.Detail != "" {
				rowText = fmt.Sprintf("%s: %s", row.Label, row.Detail)
			}
			label := widget.NewText(
				widget.TextOpts.Text(rowText, game.theme.ButtonTheme.TextFace, color.White),
			)
			sectionContainer.AddChild(label)
		}
		detailContainer.AddChild(sectionContainer)
	}

	detailScroll := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(detailContainer),
		widget.ScrollContainerOpts.Image(createScrollContainerImage()),
		widget.ScrollContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.GridLayoutData{
			MaxHeight: 0,
		})),
	)
	root.AddChild(detailScroll)

	return root
}
