//go:build gui

// Builds retained tab-body widget layouts.
// Dependencies: github.com/cjtrowbridge/apparat/internal/hud, github.com/ebitenui/ebitenui, github.com/ebitenui/ebitenui/widget.
package gui

import (
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"strings"
)

func (game *Game) rebuildUI(snapshot hud.Snapshot) {
	if game.actions.mode != "" {
		game.buildActionScreen()
		return
	}
	game.projects.workspace.layout.begin()
	if snapshot.ActiveTab().ID() != hud.TabSettings {
		game.prefs.pane = nil
	}
	if game.tabScroll != nil {
		game.tabStripScrollLeft = clampScrollLeft(game.tabScroll.ScrollLeft)
	}
	game.updateButton = nil
	game.updateDetail = nil
	game.tabScroll = nil
	game.tabButtonCount = 0
	game.tabStripContentWidth = 0
	game.tabButtonLefts = nil
	game.tabButtonRights = nil
	game.tabButtons = nil
	game.tabRadioGroup = nil
	game.verticalScrolls = nil
	game.divider = nil
	game.planning.divider = nil
	game.clampSplitWidth()
	title := snapshot.ActiveTab().Title()
	title, bottomInset := game.chatLayoutInsets(snapshot, title)
	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(&widget.Insets{
				Left:   windowMargin,
				Right:  windowMargin,
				Top:    tabTop,
				Bottom: bottomInset,
			}),
		)),
	)

	rail := game.sideRail()
	rows := []bool{false, false, true}
	contentLayout := widget.AnchorLayoutData{StretchHorizontal: true, StretchVertical: true}
	bottom := game.bottomTabs()
	if bottom {
		rows = []bool{false, true, false}
	}
	if rail != "" {
		rows = []bool{false, true}
		inset := &widget.Insets{}
		if rail == "left_rail" {
			inset.Left = railGutter
		} else {
			inset.Right = railGutter
		}
		contentLayout.Padding = inset
	}
	gallery := snapshot.ActiveTab().ID() == "stmpck"
	if gallery {
		titleRow := 0
		if rail == "" && !bottom {
			titleRow = 1
		}
		rows = append(rows[:titleRow], rows[titleRow+1:]...)
	}
	shell := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, rows),
			widget.GridLayoutOpts.Spacing(0, bodyGap),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(contentLayout)),
	)
	strip := game.buildTabStrip(snapshot)
	if rail == "" && !bottom {
		shell.AddChild(strip)
	} else if rail != "" {
		anchor := widget.AnchorLayoutData{}
		if rail == "right_rail" {
			anchor.HorizontalPosition = widget.AnchorLayoutPositionEnd
		}
		strip.GetWidget().LayoutData = anchor
		root.AddChild(strip)
	}
	if !gallery {
		game.addHeading(shell, title, snapshot.ActiveTab().ID())
	}
	game.projects.wizard.focusControl = nil
	body := game.buildActiveTabBody(snapshot)
	body.GetWidget().LayoutData = widget.GridLayoutData{MaxHeight: 0, MaxWidth: game.hudPreferredWidth()}
	shell.AddChild(body)
	if bottom {
		shell.AddChild(strip)
	}
	root.AddChild(shell)

	var old widget.Focuser
	if game.ui != nil {
		old = game.ui.GetFocusedWidget()
	}
	game.nav.body = body
	changedTab := game.nav.owner != snapshot.ActiveTab().ID()
	game.nav.owner = snapshot.ActiveTab().ID()
	game.ui = &ebitenui.UI{
		DisableDefaultFocus: true,
		Container:           root,
		PrimaryTheme:        game.theme,
	}
	game.finishWorkspaceLayout()
	a, t := game.currentChatTerminal()
	if t != nil && t.permissionPage {
		game.openChatPermissionModal(a, t)
	}
	game.restoreEditing(old)
	if changedTab && !game.nav.strip && snapshot.ActiveTab().ID() == hud.TabGlobalSession {
		game.focusGlobalPrompt()
	}
	if game.nav.active && game.ui.GetFocusedWidget() == nil {
		if changedTab || game.nav.strip {
			game.focusStrip()
		} else {
			game.nav.target = focusTarget{} // Restore after the new tree has laid out.
		}
	}
}

func (game *Game) buildTabStrip(snapshot hud.Snapshot) widget.PreferredSizeLocateableWidget {
	direction := widget.DirectionHorizontal
	if game.sideRail() != "" {
		direction = widget.DirectionVertical
	}
	row := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(direction),
			widget.RowLayoutOpts.Spacing(tabGap),
		)),
	)
	game.tabButtonCount = len(snapshot.Tabs)
	left := 0
	for index, tabData := range snapshot.Tabs {
		width := tabTargetWidth(game.hudPreferredWidth(), len(snapshot.Tabs))
		button := widget.NewButton(
			game.tabIconContent(tabData.Descriptor),
			widget.ButtonOpts.Image(tabButtonImage()),
			widget.ButtonOpts.TextPadding(&widget.Insets{}),
			widget.ButtonOpts.GraphicPadding(widget.Insets{}),
			widget.ButtonOpts.ToggleMode(),
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.MinSize(width, tabHeight),
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
			),
		)
		// Explicit tab styling avoids an inherited empty text slot beside graphics.
		button.GetWidget().SetTheme(&widget.Theme{})
		if index == snapshot.ActiveIndex {
			button.SetState(widget.WidgetChecked)
		}
		game.tabButtons = append(game.tabButtons, button)
		game.tabButtonLefts = append(game.tabButtonLefts, left)
		left += width
		game.tabButtonRights = append(game.tabButtonRights, left)
		left += tabGap
		row.AddChild(button)
	}
	contentWidth, _ := row.PreferredSize()
	game.tabStripContentWidth = contentWidth
	game.tabRadioGroup = widget.NewRadioGroup(
		widget.RadioGroupOpts.Elements(tabButtonsAsElements(game.tabButtons)...),
		widget.RadioGroupOpts.InitialElement(game.tabButtons[snapshot.ActiveIndex]),
		widget.RadioGroupOpts.ChangedHandler(func(args *widget.RadioGroupChangedEventArgs) {
			game.handleTabRadioGroupChange(args.Active)
		}),
	)
	if game.sideRail() != "" {
		return row
	}
	scroll := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(row),
		widget.ScrollContainerOpts.Image(createScrollContainerImage()),
		widget.ScrollContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(0, tabHeight),
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{MaxHeight: tabHeight, MaxWidth: game.hudPreferredWidth()}),
		),
	)
	if !game.tabStripOverflows() {
		game.tabStripScrollLeft = 0
	}
	scroll.ScrollLeft = game.tabStripScrollLeft
	game.tabScroll = scroll
	return boundPreferredWidth(scroll, game.hudPreferredWidth)
}

func tabButtonsAsElements(buttons []*widget.Button) []widget.RadioGroupElement {
	elements := make([]widget.RadioGroupElement, len(buttons))
	for index, button := range buttons {
		elements[index] = button
	}
	return elements
}

func (game *Game) handleTabRadioGroupChange(active widget.RadioGroupElement) {
	if game.syncingTabButtonStates || game.tabSelectionSuppressed() {
		game.syncTabButtonStates()
		return
	}
	for index, tabButton := range game.tabButtons {
		if active == tabButton {
			if index != game.shell.Snapshot().ActiveIndex {
				game.selectTab(index, tabButton)
			}
			return
		}
	}
}

func (game *Game) selectTab(index int, button *widget.Button) {
	for _, tabButton := range game.tabButtons {
		tabButton.SetState(widget.WidgetUnchecked)
	}
	if button != nil {
		button.SetState(widget.WidgetChecked)
	}
	_ = game.shell.SelectTab(index)
	active := game.shell.Snapshot().ActiveTab().ID()
	if active == hud.TabCluster || active == hud.TabProjects {
		q, ok := game.runtimeInfo.Queries.(interface{ RequestSync() })
		if ok {
			q.RequestSync()
		}
	}
	game.requestActiveTabVisible()
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
	if game.tabRadioGroup != nil && active < len(game.tabButtons) {
		game.tabRadioGroup.SetActive(game.tabButtons[active])
		return
	}
	for index, tabButton := range game.tabButtons {
		if index == active {
			tabButton.SetState(widget.WidgetChecked)
			continue
		}
		tabButton.SetState(widget.WidgetUnchecked)
	}
}

func (game *Game) buildActiveTabBody(snapshot hud.Snapshot) widget.PreferredSizeLocateableWidget {
	if snapshot.ActiveTab().ID() == hud.TabGlobalSession {
		return game.globalSessionBody(snapshot.ActiveTab())
	}
	if snapshot.ActiveTab().ID() == "stmpck" {
		return widget.NewContainer()
	}
	tabData := game.presentedTab(snapshot.ActiveTab())
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

	if game.prefs.pane != nil {
		scrollContainer.ScrollTop = game.prefs.pane.ScrollTop
	}
	game.prefs.pane = scrollContainer
	game.registerVerticalScroll(scrollContainer)
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

		if game.prefs.reveal != "" && game.prefs.reveal == section.ActionID {
			game.prefs.revealWidget = sectionContainer.GetWidget()
		}
		sectionContainer.AddChild(game.hudText(strings.ToUpper(section.Title)))

		if section.Description != "" {
			sectionContainer.AddChild(game.hudText(section.Description))
		}
		game.settingsCardControls(sectionContainer, section.ActionID)
		for _, row := range section.Rows {
			rowText := row.Label
			if row.Detail != "" {
				rowText = fmt.Sprintf("%s: %s", row.Label, row.Detail)
			}
			sectionContainer.AddChild(game.hudText(rowText))
		}

		if strings.EqualFold(section.Title, "updates") && game.onCheckForUpdate != nil {
			updateLabel := game.UpdateStatus()
			if updateLabel == "" {
				updateLabel = "Check for update"
			}
			updateBtn := widget.NewButton(
				widget.ButtonOpts.Text(game.updateLabel(updateLabel), game.theme.ButtonTheme.TextFace, game.theme.ButtonTheme.TextColor),
				widget.ButtonOpts.Image(game.theme.ButtonTheme.Image),
				widget.ButtonOpts.TextPadding(game.theme.ButtonTheme.TextPadding),
				widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(0, 44), widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})),
				widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
					if game.bodySelectionSuppressed() {
						return
					}
					game.requestUpdateAction()
					game.applyUpdateStatus()
				}),
			)
			game.updateButtonStatus = updateLabel
			game.updateButton = updateBtn
			sectionContainer.AddChild(updateBtn)
			game.updateDetail = game.hudText(game.UpdateStatusDetail())
			sectionContainer.AddChild(game.updateDetail)
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
					if game.bodySelectionSuppressed() {
						game.rebuildUI(game.shell.Snapshot())
						return
					}
					p := game.prefs.Preferences
					p.Overlay = args.State == widget.WidgetChecked
					game.savePreferences(p)
					game.rebuildUI(game.shell.Snapshot())
				}),
			)
			sectionContainer.AddChild(debugToggle)
		}
		content.AddChild(sectionContainer)
	}

	return content
}
