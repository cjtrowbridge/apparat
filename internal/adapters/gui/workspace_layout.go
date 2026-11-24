//go:build gui

// Keeps the active line editor and detail viewport stable when a HUD layout changes.
// TextInput owns the caret and selection, including selections within a line.
package gui

import (
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/widget"
)

type workspaceLayout struct {
	viewport   *widget.ScrollContainer
	top        float64
	mode, path string
	folders    map[string]float64
}

func (l *workspaceLayout) begin() {
	if l.viewport != nil {
		l.top = l.viewport.ScrollTop
		if l.mode == "files" {
			if l.folders == nil {
				l.folders = map[string]float64{}
			}
			l.folders[l.path] = l.top
		}
		l.viewport = nil
	}
}

func (g *Game) restoreWorkspaceScroll(scroll *widget.ScrollContainer, tab hud.Tab) {
	w := &g.projects.workspace
	i := g.selectedSectionIndex(tab)
	if tab.ID() != hud.TabProjects || w.id == "" || i < 0 || tab.Sections[i].ProjectID != w.id || g.projects.pending != nil {
		return
	}
	l := &w.layout
	key := w.path
	if w.mode == "files" {
		key = w.dir
		scroll.ScrollTop = l.folders[key]
	} else if w.mode == l.mode && key == l.path {
		scroll.ScrollTop = l.top
	}
	l.viewport = scroll
	l.mode, l.path = w.mode, key
}

func (g *Game) finishWorkspaceLayout() {
	w := &g.projects.workspace
	if w.surface != nil && (!g.inWorkspace() || w.mode != "file") {
		w.surface.ime.Cancel()
	}
}

func (g *Game) workspaceChatPane(scroll *widget.ScrollContainer) widget.PreferredSizeLocateableWidget {
	box := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1), widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true, false}))), widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(g.detailLayoutData())))
	scroll.GetWidget().LayoutData = widget.GridLayoutData{MaxHeight: 0}
	box.AddChild(scroll)
	c := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))))
	w := &g.projects.workspace
	g.multilineProjectField(c, "Message", &w.message)
	if w.attachment != "" {
		c.AddChild(g.detailText(w.attachmentLabel))
	}
	g.projectButton(c, "Post locally", func() {
		w.argument = w.message
		action := "post"
		if w.attachment != "" {
			action = "post-with-run"
		}
		g.workspaceRequest(action)
	})
	g.projectButton(c, "Attach a run", func() {
		g.workspaceGo("attachments", "runs")
	})
	box.AddChild(c)
	return boundPreferredWidth(box, g.detailPanePreferredWidth)
}

func (g *Game) clampBodyScroll() {
	for _, pane := range g.verticalScrolls {
		moveBodyScroll(pane, 0)
	}
}

func (g *Game) workspaceRow(c *widget.Container, label string, click func()) {
	width := max(80, g.detailPanePreferredWidth()-68)
	label = wrapSelectorLabel(label, *g.theme.ButtonTheme.TextFace, width)
	c.AddChild(widget.NewButton(widget.ButtonOpts.Text(label, g.theme.ButtonTheme.TextFace, g.theme.ButtonTheme.TextColor),
		widget.ButtonOpts.TextPosition(widget.TextPositionStart, widget.TextPositionCenter), widget.ButtonOpts.Image(g.theme.ButtonTheme.Image),
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(0, 44), widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			if !g.bodySelectionSuppressed() && g.projects.pending == nil {
				click()
			}
		})))
}
func (g *Game) workspaceNavButton(c *widget.Container, label string, active bool, click func()) {
	images := tabButtonImage()
	if active {
		images.Idle = images.Pressed
	}
	button := widget.NewButton(widget.ButtonOpts.Graphic(&widget.GraphicImage{Idle: emojiIcon(label)}), widget.ButtonOpts.Image(images), widget.ButtonOpts.GraphicPadding(widget.Insets{}), widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(44, 44)), widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
		if !g.bodySelectionSuppressed() && g.projects.pending == nil {
			click()
		}
	}))
	button.GetWidget().SetTheme(&widget.Theme{})
	c.AddChild(button)
}

func reattachEditor(editor *widget.TextInput) {
	// AddChild installs these four parent forwarding handlers on every attachment.
	// This input has no application subscriptions on those events.
	v := editor.GetWidget()
	v.FocusEvent = &event.Event{}
	v.ContextMenuEvent = &event.Event{}
	v.ToolTipEvent = &event.Event{}
	v.DragAndDropEvent = &event.Event{}
}
