//go:build gui

// Exercises layout rebuilds while a project draft has keyboard focus.
// Checks that scroll and edit state follow the selected file across the pane breakpoint.
package gui

import (
	"testing"

	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
)

type layoutRegistry struct{ projectStore }

func TestWorkspaceResizePreservesPresentation(t *testing.T) {
	g := NewGame()
	defer g.Close()
	g.projects.store = layoutRegistry{}
	g.projects.rows = []domain.Project{{ID: "p", Name: "workspace", Available: true}}
	selectTabThroughStrip(t, g, hud.TabProjects)
	w := &g.projects.workspace
	*w = workspacePanel{id: "p", mode: "file", path: "sample.txt", lines: []string{"draft text"}, report: "Conflict: file changed"}
	g.detailOpen[hud.TabProjects] = true
	g.selectSection(hud.TabProjects, 2)
	l := &w.layout
	editor := l.editor
	g.ui.SetFocusedWidget(editor)
	editor.SelectAll()
	l.viewport.ScrollTop = 0.45
	for _, width := range []int{640, 1024, 720} {
		g.Layout(width, 600)
		g.rebuildUI(g.shell.Snapshot())
		if g.ui.GetFocusedWidget() != editor || editor.SelectedText() != "draft text" || l.viewport == nil || l.viewport.ScrollTop != 0.45 {
			t.Fatalf("lost editor selection, focus or detail scroll at width %d", width)
		}
		if w.path != "sample.txt" || w.report != "Conflict: file changed" || w.lines[0] != "draft text" {
			t.Fatalf("resize changed draft or conflict state at width %d", width)
		}
	}
	w.path = "another.txt"
	g.rebuildUI(g.shell.Snapshot())
	if l.editor == editor || l.viewport.ScrollTop != 0 || g.ui.GetFocusedWidget() != nil {
		t.Fatal("new file inherited presentation from a different file")
	}
}
