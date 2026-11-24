//go:build gui

// Exercises layout rebuilds while a project draft has keyboard focus.
// Checks that scroll and edit state follow the selected file across the pane breakpoint.
package gui

import (
	"errors"
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
	*w = workspacePanel{id: "p", mode: "file", editing: true, path: "sample.txt", lines: []string{"draft text"}, report: "Conflict: file changed"}
	g.detailOpen[hud.TabProjects] = true
	g.selectSection(hud.TabProjects, len(projectEntrySections))
	l := &w.layout
	editor := w.surface.TextInput
	g.ui.SetFocusedWidget(editor)
	w.surface.caret = len(w.surface.body)
	l.viewport.ScrollTop = 0.45
	for _, width := range []int{640, 1024, 720} {
		g.Layout(width, 600)
		g.rebuildUI(g.shell.Snapshot())
		if g.ui.GetFocusedWidget() != editor || w.surface.body[w.surface.anchor:w.surface.caret] != "draft text" || l.viewport == nil || l.viewport.ScrollTop != 0.45 {
			t.Fatalf("lost editor selection, focus or detail scroll at width %d", width)
		}
		if w.path != "sample.txt" || w.report != "Conflict: file changed" || w.lines[0] != "draft text" {
			t.Fatalf("resize changed draft or conflict state at width %d", width)
		}
	}
	g.applyWorkspace(&workspaceReply{request: domain.WorkspaceRequest{ProjectID: w.id, Action: "read", Path: "another.txt"}, view: domain.WorkspaceView{Editing: true, Text: "next"}})
	g.rebuildUI(g.shell.Snapshot())
	if w.surface.TextInput == editor || l.viewport.ScrollTop != 0 || g.ui.GetFocusedWidget() != nil {
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

func TestOwnerTitleAndCatalogNavigation(t *testing.T) {
	g := NewGame()
	defer g.Close()
	g.serviceProjection.ProjectHostname = "pixel-4a"
	local := domain.Project{ID: "a", Name: "sample"}
	peer := domain.Project{ID: "worker-test/r", Name: "sample", RemoteHost: "trashcan0", NextCursor: "r"}
	if g.projectTitle(local) != "pixel-4a://sample" || g.projectTitle(peer) != "trashcan0://sample" {
		t.Fatal("incorrect title while identity is locked")
	}
	g.projects.rows = []domain.Project{local, peer}
	g.serviceProjection.ProjectHostname = ""
	g.runtimeInfo.DeviceName = "Pixel 4a"
	if g.projectTitle(local) != "Pixel-4a://sample" {
		t.Fatal("missing Android display name fallback")
	}
	turns := g.projects.pageTurns()
	if len(turns) != 1 || turns[0].host != "worker-test" || turns[0].cursor != "r" {
		t.Fatal(turns)
	}
	for _, width := range []int{360, 1200} {
		g.Layout(width, 640)
		for _, dst := range workspaceDestinations {
			g.openWorkspaceTab(dst.mode, "")
			if g.projects.workspace.mode != dst.mode || emojiIcon(dst.icon) == nil {
				t.Fatal(dst)
			}
		}
	}
}

func TestFailedFolderReplacesPreviousListing(t *testing.T) {
	g := NewGame()
	defer g.Close()
	w := &g.projects.workspace
	*w = workspacePanel{id: "worker-fixture/project", dir: "sub", mode: "files", cursor: "previous", entries: []domain.WorkspaceEntry{{Name: "from-parent"}}, changes: []domain.WorkspaceChange{{Path: "from-parent"}}}
	g.applyWorkspace(&workspaceReply{request: domain.WorkspaceRequest{ProjectID: w.id, Path: w.dir, Action: "list"}, err: errors.New("folder is unavailable")})
	if len(w.entries) != 0 || len(w.changes) != 0 || w.cursor != "" || w.report != "folder is unavailable" {
		t.Fatal("failed navigation retained the previous directory")
	}
}

func TestMemberOpensRemoteFileReadOnly(t *testing.T) {
	g := NewGame()
	defer g.Close()
	w := &g.projects.workspace
	w.id = "worker-fixture/project"
	g.applyWorkspace(&workspaceReply{request: domain.WorkspaceRequest{ProjectID: w.id, Action: "read", Path: "README.md"}, view: domain.WorkspaceView{ReadOnly: true, Editing: true, Text: "welcome"}})
	if w.mode != "file" || !w.readOnly || w.editing || len(w.lines) != 1 || w.lines[0] != "welcome" || g.projects.pending != nil {
		t.Fatal("member did not open the file in read-only mode")
	}
}
