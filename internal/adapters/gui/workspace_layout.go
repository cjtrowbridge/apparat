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
	editor     *widget.TextInput
	file       string
	line       int
	visible    bool
	theme      *widget.Theme
	viewport   *widget.ScrollContainer
	top        float64
	mode, path string
}

func (l *workspaceLayout) begin() {
	l.visible = false
	if l.viewport != nil {
		l.top = l.viewport.ScrollTop
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
	if w.mode == l.mode && w.path == l.path {
		scroll.ScrollTop = l.top
	}
	l.viewport = scroll
	l.mode, l.path = w.mode, w.path
}

func (g *Game) workspaceLineInput(c *widget.Container, w *workspacePanel, changed func(string)) {
	l := &w.layout
	text := w.lines[w.line]
	if l.editor == nil || l.file != w.path || l.line != w.line || l.theme != g.theme || l.editor.GetText() != text {
		l.editor = g.newProjectInput(text, changed)
		l.file, l.line, l.theme = w.path, w.line, g.theme
	} else {
		// AddChild installs these four parent forwarding handlers on every attachment.
		// This input has no application subscriptions on those events.
		v := l.editor.GetWidget()
		v.FocusEvent = &event.Event{}
		v.ContextMenuEvent = &event.Event{}
		v.ToolTipEvent = &event.Event{}
		v.DragAndDropEvent = &event.Event{}
	}
	l.visible = true
	c.AddChild(g.detailText("Edit line"), l.editor)
}

func (g *Game) finishWorkspaceLayout() {
	l := &g.projects.workspace.layout
	if !l.visible {
		l.editor = nil
	} else if l.editor.IsFocused() {
		g.ui.SetFocusedWidget(l.editor)
	}
}
