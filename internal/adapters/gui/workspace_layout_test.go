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
	*w = workspacePanel{id: "p", mode: "inline", path: "sample.txt", lines: []string{"draft text"}, report: "Conflict: file changed"}
	g.detailOpen[hud.TabProjects] = true
	g.selectSection(hud.TabProjects, len(projectEntrySections))
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

func TestDirectoryItemsIncludeDirectChangesAndMissingSubtrees(t *testing.T) {
	entries := []domain.WorkspaceEntry{{Name: "clean"}, {Name: "new"}, {Name: "edited"}}
	changes := []domain.WorkspaceChange{{Path: "new", Marker: "U"}, {Path: "edited", Marker: "C", Staged: true}, {Path: "old/gone", Marker: "D"}, {Path: "gone", Marker: "D"}}
	items := directoryItems(".", entries, changes)
	if len(items) != 5 || items[0].Name != "old" || !items[0].Directory || !items[0].missing || items[1].change.Marker != "" || items[2].change.Marker != "C" || items[3].change.Marker != "D" || items[4].change.Marker != "U" {
		t.Fatal(items)
	}
	nested := directoryItems("old", nil, changes)
	if len(nested) != 1 || nested[0].Directory || !nested[0].missing || nested[0].change.Path != "old/gone" {
		t.Fatal(nested)
	}
}
