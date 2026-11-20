//go:build gui

// Builds the selector and pins repository entry actions above its viewport.
// Uses HUD sections and EbitenUI layout widgets.
package gui

import (
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
)

const selectorListInset = 8

func projectEntryAction(kind hud.ContentKind) bool {
	return kind == cloneRepoForm || kind == initRepoForm || kind == importRepoForm
}

func (game *Game) selectorRows() *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(4),
			widget.RowLayoutOpts.Padding(&widget.Insets{Left: selectorListInset, Right: selectorListInset, Top: 8, Bottom: 8}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(game.masterListWidth(), 0),
		),
	)
}

func (game *Game) buildMasterList(tabData hud.Tab, layoutData interface{}) widget.PreferredSizeLocateableWidget {
	listContainer := game.selectorRows()
	var actions *widget.Container
	for index, section := range tabData.Sections {
		if section.IsSelectorHeading() {
			listContainer.AddChild(game.selectorHeading(section))
			continue
		}
		sectionIndex := index
		btn, subtitle := game.sectionButton(tabData, section, sectionIndex)
		if game.selectedSectionIndex(tabData) == index {
			btn.SetState(widget.WidgetChecked)
		}
		hasIcon := tabData.ID() == hud.TabProjects && (section.HasConfiguredActions || section.ContentKind == hud.ContentPipeline)
		if tabData.ID() == hud.TabProjects && projectEntryAction(section.ContentKind) {
			if actions == nil {
				actions = game.selectorRows()
			}
			row := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
				widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})))
			btn.GetWidget().LayoutData = widget.AnchorLayoutData{StretchHorizontal: true, StretchVertical: true}
			row.AddChild(btn, widget.NewGraphic(widget.GraphicOpts.Image(emojiIcon("2795")),
				widget.GraphicOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					VerticalPosition: widget.AnchorLayoutPositionCenter, Padding: &widget.Insets{Left: 12},
				}))))
			actions.AddChild(row)
			continue
		}
		if subtitle != nil || hasIcon {
			row := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
				widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})))
			btn.GetWidget().LayoutData = widget.AnchorLayoutData{StretchHorizontal: true, StretchVertical: true}
			row.AddChild(btn)
			if subtitle != nil {
				row.AddChild(subtitle)
			}
			if hasIcon {
				row.AddChild(widget.NewGraphic(widget.GraphicOpts.Image(emojiIcon("1f9e0")),
					widget.GraphicOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
						HorizontalPosition: widget.AnchorLayoutPositionEnd, VerticalPosition: widget.AnchorLayoutPositionCenter,
						Padding: &widget.Insets{Right: 12},
					}))))
			}
			listContainer.AddChild(row)
		} else {
			listContainer.AddChild(btn)
		}
	}
	scroll := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(listContainer),
		widget.ScrollContainerOpts.Image(createScrollContainerImage()),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(layoutData)),
	)
	game.registerVerticalScroll(scroll)
	if actions == nil {
		return boundPreferredWidth(scroll, game.masterPanePreferredWidth)
	}
	box := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1),
		widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true}))),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(layoutData)))
	scroll.GetWidget().LayoutData = widget.GridLayoutData{}
	box.AddChild(actions, scroll)
	return boundPreferredWidth(box, game.masterPanePreferredWidth)
}
