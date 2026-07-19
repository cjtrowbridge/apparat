//go:build gui

package gui

import (
	"fmt"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
)

func (game *Game) buildMasterDetailTab(tabData hud.Tab) widget.PreferredSizeLocateableWidget {
	if game.collapsed() {
		if game.detailOpen[tabData.ID()] {
			return game.buildCollapsedDetail(tabData)
		}
		return game.buildCollapsedList(tabData)
	}
	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(3),
			widget.GridLayoutOpts.Stretch([]bool{false, false, true}, []bool{true}),
			widget.GridLayoutOpts.Spacing(8, 0),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{StretchHorizontal: true, StretchVertical: true})),
	)
	root.AddChild(game.buildMasterList(tabData, widget.GridLayoutData{MaxHeight: 0}))
	root.AddChild(game.buildDivider())
	root.AddChild(game.buildDetailScroll(tabData, false))
	return boundPreferredWidth(root, game.hudPreferredWidth)
}

func (game *Game) buildCollapsedList(tabData hud.Tab) widget.PreferredSizeLocateableWidget {
	return game.buildMasterList(tabData, widget.AnchorLayoutData{StretchHorizontal: true, StretchVertical: true})
}

func (game *Game) buildCollapsedDetail(tabData hud.Tab) widget.PreferredSizeLocateableWidget {
	return game.buildDetailScroll(tabData, true)
}

func (game *Game) buildMasterList(tabData hud.Tab, layoutData interface{}) widget.PreferredSizeLocateableWidget {
	listContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(4),
			widget.RowLayoutOpts.Padding(&widget.Insets{Left: 8, Right: 8, Top: 8, Bottom: 8}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(game.masterListWidth(), 0),
			widget.WidgetOpts.LayoutData(layoutData),
		),
	)
	for index, section := range tabData.Sections {
		if section.IsSelectorHeading() {
			listContainer.AddChild(game.selectorHeading(section.Title))
			continue
		}
		sectionIndex := index
		btn := game.sectionButton(tabData, section.Title, sectionIndex)
		if game.selectedSectionIndex(tabData) == index {
			btn.SetState(widget.WidgetChecked)
		}
		listContainer.AddChild(btn)
	}
	scroll := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(listContainer),
		widget.ScrollContainerOpts.Image(createScrollContainerImage()),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(layoutData)),
	)
	game.registerVerticalScroll(scroll)
	return boundPreferredWidth(scroll, game.masterPanePreferredWidth)
}

func (game *Game) selectorHeading(title string) *widget.Text {
	return widget.NewText(
		widget.TextOpts.Text(strings.ToUpper(title), game.theme.LabelTheme.Face, mutedTextColor),
		widget.TextOpts.MaxWidth(float64(game.masterPanePreferredWidth()-24)),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})),
	)
}

func (game *Game) sectionButton(tabData hud.Tab, title string, sectionIndex int) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.Text(title, game.theme.ButtonTheme.TextFace, game.theme.ButtonTheme.TextColor),
		widget.ButtonOpts.Image(game.theme.ButtonTheme.Image),
		widget.ButtonOpts.TextPadding(&widget.Insets{Left: 40, Right: 12, Top: 12, Bottom: 12}),
		widget.ButtonOpts.TextPosition(widget.TextPositionStart, widget.TextPositionCenter),
		widget.ButtonOpts.ToggleMode(),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(0, tabHeight),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			if game.bodySelectionSuppressed() {
				return
			}
			game.selectSection(tabData.ID(), sectionIndex)
		}),
	)
}

func (game *Game) buildDetailScroll(tabData hud.Tab, withBack bool) widget.PreferredSizeLocateableWidget {
	detailContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(bodyGap),
			widget.RowLayoutOpts.Padding(&widget.Insets{Left: 12, Right: 12, Top: 12, Bottom: 12}),
		)),
	)
	if withBack {
		detailContainer.AddChild(game.backButton(tabData.ID()))
	}
	if tabData.Summary != "" {
		detailContainer.AddChild(game.detailText(tabData.Summary))
	}
	for _, section := range game.detailSections(tabData) {
		detailContainer.AddChild(game.buildSectionContainer(section))
	}
	scroll := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(detailContainer),
		widget.ScrollContainerOpts.Image(createScrollContainerImage()),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(game.detailLayoutData())),
	)
	game.registerVerticalScroll(scroll)
	return boundPreferredWidth(scroll, game.detailPanePreferredWidth)
}

func (game *Game) backButton(tabID hud.TabID) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.Text("<- Back", game.theme.ButtonTheme.TextFace, game.theme.ButtonTheme.TextColor),
		widget.ButtonOpts.Image(game.theme.ButtonTheme.Image),
		widget.ButtonOpts.TextPadding(game.theme.ButtonTheme.TextPadding),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(0, tabHeight),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			if game.bodySelectionSuppressed() {
				return
			}
			game.showList(tabID)
		}),
	)
}

func (game *Game) detailLayoutData() interface{} {
	if game.collapsed() {
		return widget.AnchorLayoutData{StretchHorizontal: true, StretchVertical: true}
	}
	return widget.GridLayoutData{MaxHeight: 0}
}

func (game *Game) buildSectionContainer(section hud.Section) *widget.Container {
	sectionContainer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(game.theme.PanelTheme.BackgroundImage),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(&widget.Insets{Left: 12, Right: 12, Top: 12, Bottom: 12}),
			widget.RowLayoutOpts.Spacing(4),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})),
	)
	sectionContainer.AddChild(game.detailText(strings.ToUpper(section.Title)))
	if section.Description != "" {
		sectionContainer.AddChild(game.detailText(section.Description))
	}
	for _, row := range section.Rows {
		sectionContainer.AddChild(game.detailText(rowText(row)))
	}
	for _, detail := range section.DetailSections {
		sectionContainer.AddChild(game.buildSectionContainer(detail))
	}
	return sectionContainer
}

func rowText(row hud.Row) string {
	if row.Detail == "" {
		return row.Label
	}
	return fmt.Sprintf("%s: %s", row.Label, row.Detail)
}

func (game *Game) buildDivider() widget.PreferredSizeLocateableWidget {
	return widget.NewButton(
		widget.ButtonOpts.Text("|", game.theme.ButtonTheme.TextFace, game.theme.ButtonTheme.TextColor),
		widget.ButtonOpts.Image(game.theme.ButtonTheme.Image),
		widget.ButtonOpts.TextPadding(&widget.Insets{Left: 4, Right: 4, Top: 8, Bottom: 8}),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(24, tabHeight),
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{MaxHeight: 0}),
			widget.WidgetOpts.MouseButtonPressedHandler(func(args *widget.WidgetMouseButtonPressedEventArgs) {
				game.splitDragging = true
			}),
			widget.WidgetOpts.MouseButtonReleasedHandler(func(args *widget.WidgetMouseButtonReleasedEventArgs) {
				game.splitDragging = false
			}),
			widget.WidgetOpts.CursorMoveHandler(func(args *widget.WidgetCursorMoveEventArgs) {
				if game.splitDragging {
					game.splitWidth += args.DiffX
					game.clampSplitWidth()
					game.rebuildUI(game.shell.Snapshot())
				}
			}),
		),
	)
}

func (game *Game) masterListWidth() int {
	if game.collapsed() {
		return 0
	}
	return game.splitWidth
}

func (game *Game) masterPanePreferredWidth() int {
	if game.collapsed() {
		return game.hudPreferredWidth()
	}
	return min(game.masterListWidth(), game.hudPreferredWidth())
}

func (game *Game) detailPanePreferredWidth() int {
	if game.collapsed() {
		return game.hudPreferredWidth()
	}
	return max(120, game.hudPreferredWidth()-game.masterListWidth()-40)
}

func (game *Game) selectedSectionIndex(tabData hud.Tab) int {
	index := game.selectedSections[tabData.ID()]
	if !tabData.IsSelectableSection(index) {
		return tabData.FirstSelectableSectionIndex()
	}
	return index
}

func (game *Game) detailSections(tabData hud.Tab) []hud.Section {
	if len(tabData.Sections) == 0 {
		return nil
	}
	index := game.selectedSectionIndex(tabData)
	return []hud.Section{tabData.Sections[index]}
}
