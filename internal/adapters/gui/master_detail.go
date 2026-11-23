//go:build gui

// Builds selector and content-panel widgets.
// Dependencies: github.com/cjtrowbridge/apparat/internal/hud, github.com/ebitenui/ebitenui/image, github.com/ebitenui/ebitenui/widget.
package gui

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/hud"
	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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

func (game *Game) selectorHeading(section hud.Section) *widget.Container {
	width := float64(game.masterPanePreferredWidth() - 24)
	heading := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(2),
			widget.RowLayoutOpts.Padding(&widget.Insets{Left: 8, Right: 8, Top: 6, Bottom: 2}),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})),
	)
	heading.AddChild(widget.NewText(
		widget.TextOpts.Text(strings.ToUpper(section.Title), game.theme.LabelTheme.Face, selectorColor(section.SelectorColor)),
		widget.TextOpts.MaxWidth(width),
	))
	if section.Description != "" {
		heading.AddChild(widget.NewText(
			widget.TextOpts.Text(section.Description, game.selectorDescriptionFace, mutedTextColor),
			widget.TextOpts.MaxWidth(width),
		))
	}
	return heading
}

func (game *Game) sectionButton(tabData hud.Tab, section hud.Section, sectionIndex int) (*widget.Button, *widget.Text) {
	rightPadding := 12
	if tabData.ID() == hud.TabProjects && (section.HasConfiguredActions || section.ContentKind == hud.ContentPipeline) {
		rightPadding += tabIconSize
	}
	padding := &widget.Insets{Left: 40, Right: rightPadding, Top: 12, Bottom: 12}
	if section.ActionID == "add-new-device" {
		padding.Left = 12 + tabIconSize + 8
	}
	if tabData.ID() == hud.TabProjects {
		padding.Left = 12
		if projectEntryAction(section.ContentKind) {
			padding.Left += tabIconSize + 4
		}
	}
	width := game.masterPanePreferredWidth() - 2*selectorListInset - padding.Left - padding.Right
	face := game.theme.ButtonTheme.TextFace
	label := wrapSelectorLabel(section.Title, *face, width)
	var subtitle *widget.Text
	if section.SelectorSubtitle != "" {
		face := game.selectorDescriptionFace
		lines := strings.Split(section.SelectorSubtitle, "\n")
		for i, line := range lines {
			lines[i] = wrapSelectorPath(line, *face, width)
		}
		text := strings.Join(lines, "\n")
		subtitle = widget.NewText(widget.TextOpts.Text(text, face, mutedTextColor),
			widget.TextOpts.MaxWidth(float64(width)),
			widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				VerticalPosition: widget.AnchorLayoutPositionEnd,
				Padding:          &widget.Insets{Left: padding.Left, Right: rightPadding, Bottom: padding.Bottom},
			})))
		subtitle.Validate()
		_, height := subtitle.PreferredSize()
		padding.Bottom += height + 3
	}
	return widget.NewButton(
		widget.ButtonOpts.Text(label, face, game.theme.ButtonTheme.TextColor),
		widget.ButtonOpts.Image(selectorButtonImage(section.SelectorColor)),
		widget.ButtonOpts.TextPadding(padding),
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
	), subtitle
}

func (game *Game) buildDetailScroll(tabData hud.Tab, withBack bool) widget.PreferredSizeLocateableWidget {
	detailContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(bodyGap),
			widget.RowLayoutOpts.Padding(&widget.Insets{Left: 12, Right: 12, Top: 12, Bottom: 12}),
		)),
	)
	child := tabData.ID() == hud.TabCluster && game.clusterTop() != nil
	if child {
		game.clusterContent(detailContainer)
	} else {
		selected := game.selectedSectionIndex(tabData)
		peer := tabData.ID() == hud.TabCluster && game.lan.selected != ""
		if withBack && !peer && selected >= 0 && !(tabData.ID() == hud.TabProjects && game.projects.workspace.id != "" && tabData.Sections[selected].ProjectID == game.projects.workspace.id && game.projects.workspace.mode != "registration") {
			detailContainer.AddChild(game.backButton(tabData.ID()))
		}
		if selected >= 0 && tabData.Summary != "" && !peer {
			detailContainer.AddChild(game.detailText(tabData.Summary))
		}
		for _, section := range game.detailSections(tabData) {
			if section.ActionID == "add-new-device" {
				game.discoveryContent(detailContainer)
			} else {
				detailContainer.AddChild(game.buildSectionContainer(section))
				check := game.workerAction(section.ActionID, game.runtimeInfo.CheckWorker)
				if tabData.ID() == hud.TabCluster && check != nil {
					game.projectButton(detailContainer, "Check my access", check)
				}
				pool := game.workerPage(section.ActionID, "work", false)
				if tabData.ID() == hud.TabCluster && pool != nil {
					game.projectButton(detailContainer, "Pool tasks", pool)
				}
				manage := game.workerPage(section.ActionID, "machine", true)
				if tabData.ID() == hud.TabCluster && manage != nil {
					game.projectButton(detailContainer, "Worker settings", manage)
				}
			}
		}
	}
	scroll := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(detailContainer),
		widget.ScrollContainerOpts.Image(createScrollContainerImage()),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(game.detailLayoutData())),
	)
	game.registerVerticalScroll(scroll)
	if child {
		game.cluster.pane = scroll
		scroll.ScrollTop = game.clusterTop().scroll
	}
	game.restoreWorkspaceScroll(scroll, tabData)
	if tabData.ID() == hud.TabProjects && game.inWorkspace() && game.projects.workspace.mode == "conversation" {
		return game.workspaceChatPane(scroll)
	}
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
	if section.ProjectID != "" && game.projects.workspace.id == section.ProjectID && game.projects.workspace.mode != "registration" {
		game.projectControls(sectionContainer, section)
		return sectionContainer
	}
	sectionContainer.AddChild(game.detailText(strings.ToUpper(section.Title)))
	if section.Description != "" {
		sectionContainer.AddChild(game.detailText(section.Description))
	}
	for _, row := range section.Rows {
		sectionContainer.AddChild(game.detailText(rowText(row)))
	}
	if section.ContentKind == hud.ContentResearchContribution {
		sectionContainer.AddChild(game.researchContributionTable(section))
	}
	for _, detail := range section.DetailSections {
		sectionContainer.AddChild(game.buildSectionContainer(detail))
	}
	if section.ProjectID != "" || projectEntryAction(section.ContentKind) {
		game.projectControls(sectionContainer, section)
		return sectionContainer
	}
	switch section.ContentKind {
	case hud.ContentChat:
		sectionContainer.AddChild(game.chatComposer())
	case hud.ContentProject, hud.ContentPipeline:
		sectionContainer.AddChild(game.workspaceControls(section.ContentKind))
	}
	return sectionContainer
}

func (game *Game) chatComposer() *widget.Container {
	composer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{true}),
			widget.GridLayoutOpts.Spacing(6, 0),
		)),
	)
	input := game.editorForLayout("mock reply", func() *widget.TextInput {
		return widget.NewTextInput(
			widget.TextInputOpts.Face(game.theme.ButtonTheme.TextFace),
			widget.TextInputOpts.Placeholder("Type a mock reply…"),
			widget.TextInputOpts.SubmitOnEnter(false),
			widget.TextInputOpts.Image(&widget.TextInputImage{Idle: uiimage.NewNineSliceColor(panelColor), Highlight: uiimage.NewNineSliceColor(activeTabColor)}),
			widget.TextInputOpts.Color(&widget.TextInputColor{Idle: color.White, Caret: color.White}),
			widget.TextInputOpts.Padding(&widget.Insets{Left: 10, Right: 10, Top: 8, Bottom: 8}),
			widget.TextInputOpts.WidgetOpts(widget.WidgetOpts.MinSize(0, tabHeight), widget.WidgetOpts.LayoutData(widget.GridLayoutData{MaxHeight: tabHeight})),
		)
	})
	composer.AddChild(input)
	composer.AddChild(widget.NewButton(
		widget.ButtonOpts.Text("Send", game.theme.ButtonTheme.TextFace, game.theme.ButtonTheme.TextColor),
		widget.ButtonOpts.Image(game.theme.ButtonTheme.Image),
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(88, tabHeight), widget.WidgetOpts.LayoutData(widget.GridLayoutData{MaxHeight: tabHeight})),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {}),
	))
	return composer
}

func (game *Game) workspaceControls(kind hud.ContentKind) *widget.Container {
	labels := []string{"Git", "Chat"}
	if kind == hud.ContentPipeline {
		labels = append(labels, "Run")
	}
	controls := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionHorizontal), widget.RowLayoutOpts.Spacing(6))))
	for _, label := range labels {
		controls.AddChild(widget.NewButton(
			widget.ButtonOpts.Text(label, game.theme.ButtonTheme.TextFace, game.theme.ButtonTheme.TextColor),
			widget.ButtonOpts.Image(game.theme.ButtonTheme.Image),
			widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(72, 44)),
			widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {}),
		))
	}
	return controls
}

func rowText(row hud.Row) string {
	if row.Detail == "" {
		return row.Label
	}
	return fmt.Sprintf("%s: %s", row.Label, row.Detail)
}

func (game *Game) buildDivider() widget.PreferredSizeLocateableWidget {
	divider := widget.NewButton(
		widget.ButtonOpts.Text("|", game.theme.ButtonTheme.TextFace, game.theme.ButtonTheme.TextColor),
		widget.ButtonOpts.Image(game.theme.ButtonTheme.Image),
		widget.ButtonOpts.TextPadding(&widget.Insets{Left: 4, Right: 4, Top: 8, Bottom: 8}),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(24, tabHeight),
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{MaxHeight: 0}),
		),
	)
	game.divider = divider.GetWidget()
	return divider
}

func (game *Game) updateDividerDrag(x, y int) {
	if game.divider == nil || !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		game.releaseGesture(gestureDivider)
		return
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) &&
		imagePoint(x, y).In(game.divider.Rect) &&
		!game.pointInDebugOverlay(x, y) && !game.pointInPTT(x, y) &&
		game.claimGesture(gestureDivider, false, 0) {
		game.splitGrabOffset = x - game.splitWidth
	}
	if game.gesture.Owner != gestureDivider {
		return
	}
	previous := game.splitWidth
	game.splitWidth = x - game.splitGrabOffset
	game.clampSplitWidth()
	if game.splitWidth == previous {
		return
	}
	scrolls := game.verticalScrolls
	game.rebuildUI(game.shell.Snapshot())
	for i, scroll := range game.verticalScrolls {
		scroll.ScrollTop = scrolls[i].ScrollTop
	}
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
	index, selected := game.selectedSections[tabData.ID()]
	if !selected {
		return -1
	}
	if !tabData.IsSelectableSection(index) {
		return -1
	}
	return index
}
