//go:build gui

// Presents the local workspace through the Projects detail pane.
// Requests and results carry the selected project ID across the async boundary.
package gui

import (
	"context"
	"path"
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
	id, dir, path, base, report, argument, runID string
	mode, confirmation                           string
	entries                                      []domain.WorkspaceEntry
	lines                                        []string
	line, page                                   int
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
	q := domain.WorkspaceRequest{ProjectID: w.id, Action: action, Path: file, Base: w.base, Text: strings.Join(w.lines, "\n"), Argument: w.argument}
	switch action {
	case "list":
		q.Path = w.dir
	case "create":
		q.Path = path.Join(w.dir, w.argument)
	case "save", "post", "run":
		q.ID, _ = domain.NewID(time.Now(), nil)
	case "artifact":
		q.ID = strings.TrimSpace(w.argument)
		if q.ID == "" {
			q.ID = w.runID
		}
	}
	w.confirmation = ""
	g.projectWork(func(ctx context.Context) projectResult {
		view, err := store.Workspace(ctx, q)
		return projectResult{workspace: &workspaceReply{q, view, err}}
	})
}
func (g *Game) applyWorkspace(r *workspaceReply) {
	w := &g.projects.workspace
	if w.id != r.request.ProjectID {
		return
	}
	if r.err != nil {
		w.report = r.err.Error()
		return
	}
	q, v := r.request, r.view
	w.report = v.Report
	if v.Editing {
		w.lines = strings.Split(v.Text, "\n")
		w.path = q.Path
		w.base = v.Base
		w.line = min(w.line, len(w.lines)-1)
		w.mode = "file"
	}
	switch q.Action {
	case "list":
		w.entries = v.Entries
		w.mode = "files"
		w.path = ""
		w.lines = nil
	case "move", "delete":
		w.dir = path.Dir(q.Path)
		g.workspaceRequest("list")
	case "run":
		w.runID = q.ID
		w.argument = q.ID
	}
}
func (g *Game) workspaceControlsFor(c *widget.Container, id string) bool {
	w := &g.projects.workspace
	if w.id != id {
		g.projectButton(c, "Open workspace", func() {
			*w = workspacePanel{id: id, dir: ".", mode: "files"}
			g.workspaceRequest("list")
		})
		return false
	}
	if w.mode == "menu" {
		g.projectButton(c, "Registration details", func() {
			*w = workspacePanel{}
			g.layoutDirty = true
		})
		for _, mode := range []string{"Files", "Git", "Tasks", "Chat"} {
			g.projectButton(c, mode, func() {
				w.mode = strings.ToLower(mode)
				w.report = ""
				w.confirmation = ""
				switch mode {
				case "Files":
					g.workspaceRequest("list")
				case "Git":
					g.workspaceRequest("status")
				case "Tasks":
					g.workspaceRequest("tasks")
				case "Chat":
					g.workspaceRequest("chat")
				}
			})
		}
		return true
	}
	g.projectButton(c, "Workspace menu", func() {
		w.mode = "menu"
		g.layoutDirty = true
	})
	if w.report != "" {
		c.AddChild(g.detailText(w.report))
	}
	switch w.mode {
	case "files":
		g.workspaceBrowser(c, w)
	case "file":
		g.workspaceEditor(c, w)
	case "git":
		g.projectField(c, "Branch, commit ID or commit message", &w.argument)
		g.workspaceActions(c, "Status", "status", "History", "history", "Commit detail", "commit-detail", "Transactions", "transactions", "Branches", "branches", "Switch branch", "switch")
		if w.path != "" {
			c.AddChild(g.detailText(w.path))
			g.workspaceActions(c, "Diff against HEAD", "diff", "Stage file", "stage", "Unstage file", "unstage", "Commit staged file", "commit")
		}
	case "tasks":
		g.projectField(c, "Task ID or previous run ID", &w.argument)
		g.workspaceActions(c, "List Tasks", "tasks", "Run Task", "run", "Run history", "runs", "Open verified artifact", "artifact")
	case "chat":
		g.projectField(c, "Project message", &w.argument)
		g.workspaceActions(c, "Save pending message", "post", "Message history", "chat")
	}
	return true
}
func (g *Game) workspaceActions(c *widget.Container, labels ...string) {
	w := &g.projects.workspace
	for i := 0; i < len(labels); i += 2 {
		label, action := labels[i], labels[i+1]
		confirm := false
		switch action {
		case "delete", "move", "switch", "commit", "discard", "rebase", "run", "save":
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
	c.AddChild(g.detailText("Folder · " + w.dir))
	if w.dir != "." {
		g.projectButton(c, "Parent folder", func() {
			w.dir = path.Dir(w.dir)
			w.page = 0
			g.workspaceRequest("list")
		})
	}
	for _, e := range w.entries[min(w.page, len(w.entries)):min(w.page+12, len(w.entries))] {
		name := e.Name
		if e.Directory {
			name += "/"
		}
		g.projectButton(c, name, func() {
			dest := path.Join(w.dir, e.Name)
			if e.Directory {
				w.dir = dest
				w.page = 0
				g.workspaceRequest("list")
			} else {
				w.path = dest
				w.base = ""
				w.line = 0
				g.workspaceRequest("read")
			}
		})
	}
	if w.page > 0 {
		g.projectButton(c, "Previous files", func() {
			w.page = max(0, w.page-12)
			g.layoutDirty = true
		})
	}
	if w.page+12 < len(w.entries) {
		g.projectButton(c, "More files", func() {
			w.page += 12
			g.layoutDirty = true
		})
	}
	g.projectField(c, "New file name", &w.argument)
	g.workspaceActions(c, "Create empty text file", "create")
}
func (g *Game) workspaceEditor(c *widget.Container, w *workspacePanel) {
	c.AddChild(g.detailText(w.path))
	id, file, line := w.id, w.path, w.line
	g.projectInput(c, "Edit line", w.lines[line], func(text string) {
		if w.id != id || w.path != file || w.mode != "file" || w.line != line || w.lines[line] == text {
			return
		}
		w.lines[line] = text
		g.keepEditingDraft(w)
	})
	if w.line > 0 {
		g.projectButton(c, "Previous line", func() {
			w.line--
			g.layoutDirty = true
		})
	}
	if w.line+1 < len(w.lines) {
		g.projectButton(c, "Next line", func() {
			w.line++
			g.layoutDirty = true
		})
	}
	g.projectButton(c, "Insert line below", func() {
		w.line++
		w.lines = append(w.lines, "")
		copy(w.lines[w.line+1:], w.lines[w.line:])
		w.lines[w.line] = ""
		g.keepEditingDraft(w)
		g.layoutDirty = true
	})
	g.projectButton(c, "Remove line from draft", func() {
		w.lines = append(w.lines[:w.line], w.lines[w.line+1:]...)
		if len(w.lines) == 0 {
			w.lines = []string{""}
		}
		w.line = min(w.line, len(w.lines)-1)
		g.keepEditingDraft(w)
		g.layoutDirty = true
	})
	g.workspaceActions(c, "Keep draft", "draft", "Apply draft to file", "save", "Compare checkout text", "compare", "Use current file as base", "rebase", "Discard saved draft", "discard", "Diff against HEAD", "diff", "Stage file", "stage", "Unstage file", "unstage")
	g.projectField(c, "Commit message or destination path", &w.argument)
	g.workspaceActions(c, "Commit staged file", "commit", "Move file", "move", "Delete file", "delete")
	c.AddChild(g.detailText(strings.Join(w.lines, "\n")))
}

func (g *Game) keepEditingDraft(w *workspacePanel) {
	store, ok := g.runtimeInfo.Queries.(workspaceStore)
	if !ok {
		return
	}
	q := domain.WorkspaceRequest{ProjectID: w.id, Action: "draft", Path: w.path, Base: w.base, Text: strings.Join(w.lines, "\n")}
	_, err := store.Workspace(g.projects.ctx, q)
	if err != nil {
		w.report = "Draft could not be saved: " + err.Error()
		g.layoutDirty = true
	}
}
