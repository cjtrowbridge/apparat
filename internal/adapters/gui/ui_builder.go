//go:build gui

package gui

import (
	"fmt"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
)

// rebuildUI constructs the complete EbitenUI widget tree from the current HUD snapshot.
// In the future, this should intelligently update existing widgets instead of rebuilding
// from scratch, but for this initial migration, rebuilding ensures layout correctness.
func (game *Game) rebuildUI(snapshot hud.Snapshot) {
	game.updateButton = nil
	game.tabScroll = nil
	game.tabButtonCount = 0
	game.tabButtons = nil
	game.clampSplitWidth()
	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(&widget.Insets{
				Left:   windowMargin,
				Right:  windowMargin,
				Top:    tabTop,
				Bottom: diagnosticsHeight,
			}),
		)),
	)

	shell := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true}),
			widget.GridLayoutOpts.Spacing(0, bodyGap),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true,
			StretchVertical:   true,
		})),
	)
	shell.AddChild(game.buildTabStrip(snapshot))
	body := game.buildActiveTabBody(snapshot)
	body.GetWidget().LayoutData = widget.GridLayoutData{MaxHeight: 0, MaxWidth: game.hudPreferredWidth()}
	shell.AddChild(body)
	root.AddChild(shell)

	game.ui = &ebitenui.UI{
		Container:    root,
		PrimaryTheme: game.theme,
	}
	game.requestActiveTabVisible()
	game.ensureActiveTabVisible()
}

func (game *Game) buildTabStrip(snapshot hud.Snapshot) widget.PreferredSizeLocateableWidget {
	row := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(tabGap),
		)),
	)
	game.tabButtonCount = len(snapshot.Tabs)
	for index, tabData := range snapshot.Tabs {
		label := fmt.Sprintf("%s %s", tabData.Descriptor.Glyph, tabData.Title())
		tabIndex := index
		button := widget.NewButton(
			widget.ButtonOpts.Text(label, game.theme.ButtonTheme.TextFace, game.theme.ButtonTheme.TextColor),
			widget.ButtonOpts.Image(game.theme.ButtonTheme.Image),
			widget.ButtonOpts.TextPadding(game.theme.ButtonTheme.TextPadding),
			widget.ButtonOpts.ToggleMode(),
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.MinSize(tabButtonWidth(label), tabHeight),
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
			),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
				if game.tabSelectionSuppressed() {
					game.syncTabButtonStates()
					return
				}
				game.selectTab(tabIndex, args.Button)
			}),
			widget.ButtonOpts.StateChangedHandler(func(args *widget.ButtonChangedEventArgs) {
				if game.tabSelectionSuppressed() {
					game.syncTabButtonStates()
				}
			}),
		)
		if index == snapshot.ActiveIndex {
			button.SetState(widget.WidgetChecked)
		}
		game.tabButtons = append(game.tabButtons, button)
		row.AddChild(button)
	}
	scroll := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(row),
		widget.ScrollContainerOpts.Image(createScrollContainerImage()),
		widget.ScrollContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(0, tabHeight),
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{MaxHeight: tabHeight, MaxWidth: game.hudPreferredWidth()}),
		),
	)
	game.tabScroll = scroll
	return boundPreferredWidth(scroll, game.hudPreferredWidth)
}

func (game *Game) selectTab(index int, button *widget.Button) {
	for _, tabButton := range game.tabButtons {
		tabButton.SetState(widget.WidgetUnchecked)
	}
	if button != nil {
		button.SetState(widget.WidgetChecked)
	}
	_ = game.shell.SelectTab(index)
	game.rebuildUI(game.shell.Snapshot())
}

func (game *Game) syncTabButtonStates() {
	if game.syncingTabButtonStates {
		return
	}
	game.syncingTabButtonStates = true
	defer func() {
		game.syncingTabButtonStates = false
	}()
	active := game.shell.Snapshot().ActiveIndex
	for index, tabButton := range game.tabButtons {
		if index == active {
			tabButton.SetState(widget.WidgetChecked)
			continue
		}
		tabButton.SetState(widget.WidgetUnchecked)
	}
}

func tabButtonWidth(label string) int {
	width := 42 + len([]rune(label))*10
	if width < 130 {
		return 130
	}
	return width
}

func (game *Game) buildActiveTabBody(snapshot hud.Snapshot) widget.PreferredSizeLocateableWidget {
	tabData := snapshot.ActiveTab()
	if tabData.Descriptor.ID == hud.TabSettings {
		return game.buildSettingsTab(tabData)
	}
	return game.buildMasterDetailTab(tabData)
}

func (game *Game) buildSettingsTab(tabData hud.Tab) widget.PreferredSizeLocateableWidget {
	content := game.buildSettingsContent(tabData)
	scrollContainer := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(content),
		widget.ScrollContainerOpts.Image(createScrollContainerImage()),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true,
			StretchVertical:   true,
		})),
	)

	return boundPreferredWidth(scrollContainer, game.hudPreferredWidth)
}

func (game *Game) buildSettingsContent(tabData hud.Tab) *widget.Container {
	content := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(bodyGap),
			widget.RowLayoutOpts.Padding(&widget.Insets{Left: 12, Right: 12, Top: 12, Bottom: 12}),
		)),
	)

	if tabData.Summary != "" {
		content.AddChild(game.hudText(tabData.Summary))
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

		sectionContainer.AddChild(game.hudText(strings.ToUpper(section.Title)))

		if section.Description != "" {
			sectionContainer.AddChild(game.hudText(section.Description))
		}

		for _, row := range section.Rows {
			rowText := row.Label
			if row.Detail != "" {
				rowText = fmt.Sprintf("%s: %s", row.Label, row.Detail)
			}
			sectionContainer.AddChild(game.hudText(rowText))
		}

		if strings.ToLower(section.Title) == "updates" {
			updateLabel := game.UpdateStatus()
			if updateLabel == "" {
				updateLabel = "Check for update"
			}
			updateBtn := widget.NewButton(
				widget.ButtonOpts.Text(updateLabel, game.theme.ButtonTheme.TextFace, game.theme.ButtonTheme.TextColor),
				widget.ButtonOpts.Image(game.theme.ButtonTheme.Image),
				widget.ButtonOpts.TextPadding(game.theme.ButtonTheme.TextPadding),
				widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(0, 44)),
				widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
					game.SetUpdateStatus("Checking...")
					args.Button.SetText("Checking...")
					if game.onCheckForUpdate == nil || !game.onCheckForUpdate() {
						game.SetUpdateStatus("Update unavailable")
						args.Button.SetText("Update unavailable")
					}
				}),
			)
			game.updateButton = updateBtn
			sectionContainer.AddChild(updateBtn)
		}
		if strings.ToLower(section.Title) == "diagnostics" {
			debugLabel := "Open Debug UI overlay"
			if game.debugOverlayOpen {
				debugLabel = "Close Debug UI overlay"
			}
			debugToggle := widget.NewCheckbox(
				widget.CheckboxOpts.Text(debugLabel, game.theme.ButtonTheme.TextFace, game.theme.LabelTheme.Color),
				widget.CheckboxOpts.Image(game.theme.CheckboxTheme.Image),
				widget.CheckboxOpts.InitialState(checkedState(game.debugOverlayOpen)),
				widget.CheckboxOpts.WidgetOpts(widget.WidgetOpts.MinSize(0, 44)),
				widget.CheckboxOpts.StateChangedHandler(func(args *widget.CheckboxChangedEventArgs) {
					game.debugOverlayOpen = args.State == widget.WidgetChecked
					game.rebuildUI(game.shell.Snapshot())
				}),
			)
			sectionContainer.AddChild(debugToggle)
		}
		content.AddChild(sectionContainer)
	}

	return content
}

func checkedState(checked bool) widget.WidgetState {
	if checked {
		return widget.WidgetChecked
	}
	return widget.WidgetUnchecked
}
