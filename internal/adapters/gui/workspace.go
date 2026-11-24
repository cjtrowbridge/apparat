//go:build gui

// Presents the local workspace through the Projects detail pane.
// Requests and results carry the selected project ID across the async boundary.
package gui

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/ebitenui/ebitenui/widget"
)

type workspaceStore interface {
	Workspace(context.Context, domain.WorkspaceRequest) (domain.WorkspaceView, error)
}
type workspaceReply struct {
	request domain.WorkspaceRequest
	view    domain.WorkspaceView
	err     error
}
type workspacePanel struct {
	cursor                                       string
	virtualDirs                                  map[string]bool
	outputDetails                                string
	layout                                       workspaceLayout
	thread                                       string
	choices                                      []domain.WorkspaceChoice
	choiceAction                                 string
	selected                                     domain.WorkspaceChoice
	attachment, attachmentLabel, message         string
	readOnly, editing                            bool
	change                                       domain.WorkspaceChange
	changes                                      []domain.WorkspaceChange
	id, dir, path, base, report, argument, runID string
	mode, confirmation                           string
	entries                                      []domain.WorkspaceEntry
	lines                                        []string
	page                                         int
	originalBase, original, compared             string
	surface                                      *documentSurface
}

func (g *Game) workspaceRequest(action string) {
	w := &g.projects.workspace
	store, ok := g.runtimeInfo.Queries.(workspaceStore)
	if !ok {
		return
	}
	file := w.path
	if file == "" {
		file = w.dir
	}
	q := domain.WorkspaceRequest{Approved: action == "run", ProjectID: w.id, Action: action, Path: file, Base: w.base, Thread: w.thread, Text: strings.Join(w.lines, "\n"), Argument: w.argument}
	if action == "post-with-run" {
		q.Action = "post"
		q.RelatedRun = w.attachment
	}
	switch action {
	case "list":
		q.Argument = ""
		q.Path = w.dir
	case "rebase":
		q.Argument = w.compared
	case "create", "mkdir":
		q.Path = path.Join(w.dir, w.argument)
	case "save", "post", "post-with-run", "run":
		q.ID, _ = domain.NewID(time.Now(), nil)
	case "artifact", "delete-artifact":
		q.ID = strings.TrimSpace(w.argument)
		if action == "delete-artifact" {
			q.ID = w.runID
		}
		if q.ID == "" {
			q.ID = w.runID
		}
	}
	if action == "next-directory" {
		q.Action = "list"
		q.Argument = w.cursor
		q.Path = w.dir
	}
	if strings.HasPrefix(action, "terminal-") {
		q.ID = w.runID
	}
	virtual := action == "list" && w.virtualDirs[q.Path]
	w.confirmation = ""
	g.projectWork(func(ctx context.Context) projectResult {
		var view domain.WorkspaceView
		var err error
		if !virtual {
			view, err = store.Workspace(ctx, q)
		}
		result := projectResult{workspace: &workspaceReply{q, view, err}}
		switch q.Action {
		case "commit", "switch", "save", "create", "delete", "move":
			if err == nil && g.projects.store != nil && (q.Action == "switch" || q.Action == "commit") {
				result.rows, err = g.projects.store.Projects(ctx)
				result.reload = err == nil
				result.defaultAccount, _ = g.projects.store.ProjectDefault(ctx)
			}
		}
		return result
	})
}
func (g *Game) applyWorkspace(r *workspaceReply) {
	w := &g.projects.workspace
	if w.id != r.request.ProjectID {
		return
	}
	if r.err != nil {
		if r.request.Action == "list" || r.request.Action == "read" && errors.Is(r.err, httpapi.ErrDenied) {
			g.stopFolderCheck()
			w.entries = nil
			w.changes = nil
			w.cursor = ""
			if errors.Is(r.err, httpapi.ErrDenied) {
				w.lines = nil
				w.base = ""
				w.path = ""
				w.mode = "files"
				w.readOnly = true
			}
		}
		w.report = r.err.Error()
		return
	}
	q, v := r.request, r.view
	w.report = v.Report
	if v.Editing {
		w.readOnly = v.ReadOnly
		w.lines = strings.Split(v.Text, "\n")
		w.path = q.Path
		w.base = v.Base
		w.mode = "file"
	}
	switch q.Action {
	case "terminal-list":
		w.choices = v.Choices
		w.cursor = v.Cursor
	case "terminal-history":
		w.cursor = v.Cursor
	case "terminal-start", "terminal-stop", "terminal-take", "terminal-input", "terminal-screen", "terminal-resize":
		w.mode = "terminal-session"
		w.selected.State = v.Detail
		w.message = ""
	case "changes":
		w.changes = v.Changes
	case "review":
		w.base = v.Base
	case "read", "create", "discard":
		if w.surface != nil {
			w.surface.ime.Cancel()
		}
		w.editing = false
		w.compared = ""
		w.surface = nil
	case "compare":
		w.compared = v.ComparedBase
	case "rebase":
		w.compared = ""
	case "stage", "stage-deletion", "unstage", "commit", "switch", "resolve", "delete":
		g.workspaceGo("changes", "changes")
	case "threads", "tasks", "runs":
		w.cursor = v.Cursor
		w.choices = v.Choices
	case "post":
		for i, p := range g.projects.rows {
			if p.ID == q.ProjectID {
				thread := q.Thread
				if thread == "" {
					thread = "General"
				}
				item := domain.WorkspaceChoice{ID: thread, Label: thread}
				if !slices.Contains(p.Threads, item) {
					g.projects.rows[i].Threads = append(p.Threads, item)
				}
				break
			}
		}
		w.argument = ""
		w.message = ""
		w.attachment = ""
	case "artifact":
		w.outputDetails = v.Detail
		w.mode = "output"
		w.runID = q.ID
	case "delete-artifact":
		g.workspaceGo("runs", "runs")
	case "save":
		w.editing = false
		if w.surface != nil {
			w.surface.ime.Cancel()
		}
		w.surface = nil
		w.dir = path.Dir(q.Path)
		g.workspaceGo("files", "list")
	case "list":
		if q.Argument == "" {
			w.changes = nil
			w.entries = v.Entries
			w.page = 0
			g.startDirectoryStatus(q)
		} else {
			w.entries = v.Entries
			w.page = 0
		}
		w.cursor = v.Cursor
		w.mode = "files"
		w.path = ""
		w.lines = nil
	case "move", "mkdir":
		w.dir = path.Dir(q.Path)
		g.workspaceRequest("list")
	case "run":
		w.runID = q.ID
		w.argument = q.ID
		w.selected.State = "completed"
		w.selected.Detail = ""
		w.mode = "result"
	}
}
func (g *Game) workspaceControlsFor(c *widget.Container, id string) bool {
	w := &g.projects.workspace
	if w.id != id || w.mode == "registration" {
		return false
	}
	g.workspaceScreen(c, w)
	return true
}
func (g *Game) workspaceActions(c *widget.Container, labels ...string) {
	w := &g.projects.workspace
	for i := 0; i < len(labels); i += 2 {
		label, action := labels[i], labels[i+1]
		confirm := false
		switch action {
		case "terminal-stop", "terminal-take", "delete-artifact", "delete", "move", "switch", "commit", "discard", "rebase", "resolve", "run", "save":
			confirm = true
		}
		if confirm && w.confirmation == action {
			label = "Confirm " + label
		}
		g.projectButton(c, label, func() {
			if confirm && w.confirmation != action {
				w.confirmation = action
				g.layoutDirty = true
				return
			}
			g.workspaceRequest(action)
		})
	}
}
func (g *Game) workspaceBrowser(c *widget.Container, w *workspacePanel) {
	g.directoryListing(c, w)
}
func (g *Game) keepEditingDraft(w *workspacePanel) bool {
	store, ok := g.runtimeInfo.Queries.(workspaceStore)
	if !ok {
		return false
	}
	q := domain.WorkspaceRequest{ProjectID: w.id, Action: "draft", Path: w.path, Base: w.base, Thread: w.thread, Text: strings.Join(w.lines, "\n")}
	_, err := store.Workspace(g.projects.ctx, q)
	if err != nil {
		w.report = "Draft could not be saved: " + err.Error()
		g.layoutDirty = true
	}
	return err == nil
}
