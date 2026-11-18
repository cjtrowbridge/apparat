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
	"github.com/cjtrowbridge/apparat/internal/hud"
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
	layout                                       workspaceLayout
	thread                                       string
	choices                                      []domain.WorkspaceChoice
	choiceAction                                 string
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
	q := domain.WorkspaceRequest{Approved: action == "run", ProjectID: w.id, Action: action, Path: file, Base: w.base, Thread: w.thread, Text: strings.Join(w.lines, "\n"), Argument: w.argument}
	if action == "post-with-run" {
		q.Action = "post"
		q.RelatedRun = w.runID
	}
	switch action {
	case "list":
		q.Path = w.dir
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
	w.confirmation = ""
	g.projectWork(func(ctx context.Context) projectResult {
		view, err := store.Workspace(ctx, q)
		result := projectResult{workspace: &workspaceReply{q, view, err}}
		switch q.Action {
		case "commit", "switch", "save", "create", "delete", "move":
			if err == nil && g.projects.store != nil {
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
	case "threads":
		w.choices = v.Choices
	case "artifact":
		w.runID = q.ID
	case "tasks", "runs":
		w.choices = v.Choices
		w.choiceAction = "run"
		if q.Action == "runs" {
			w.choiceAction = "artifact"
		}
	case "list":
		w.entries = v.Entries
		w.mode = "files"
		w.path = ""
		w.lines = nil
	case "delete":
		w.mode = "git"
		w.base = "deleted"
		w.lines = nil
		w.report = "File deleted from the checkout. Stage its deletion to include it in a commit."
	case "move", "mkdir":
		w.dir = path.Dir(q.Path)
		g.workspaceRequest("list")
	case "run":
		w.runID = q.ID
		w.argument = q.ID
		w.choices = []domain.WorkspaceChoice{{ID: q.ID, Label: "Open verified artifact"}}
		w.choiceAction = "artifact"
	}
}
func (g *Game) workspaceControlsFor(c *widget.Container, id string) bool {
	w := &g.projects.workspace
	if w.id != id {
		g.projectButton(c, "Open workspace", func() {
			*w = workspacePanel{id: id, dir: ".", mode: "files"}
			g.detailOpen[hud.TabProjects] = true
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
				w.argument = ""
				w.choices = nil
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
					if w.thread == "" {
						w.thread = "General"
					}
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
			if w.base == "deleted" {
				g.workspaceActions(c, "Stage deletion", "stage-deletion", "Unstage deletion", "unstage", "Commit staged deletion", "commit")
			} else {
				g.workspaceActions(c, "Diff against HEAD", "diff", "Stage file", "stage", "Unstage file", "unstage", "Commit staged file", "commit")
			}
		}
	case "tasks":
		for _, item := range w.choices {
			label := item.Label
			if w.choiceAction == "run" {
				label = "Run " + label
				if w.confirmation == item.ID {
					label = "Confirm " + label
				}
			}
			g.projectButton(c, label, func() {
				if w.choiceAction == "run" && w.confirmation != item.ID {
					w.confirmation = item.ID
					g.layoutDirty = true
					return
				}
				w.argument = item.ID
				g.workspaceRequest(w.choiceAction)
			})
		}
		g.workspaceActions(c, "List Tasks", "tasks", "Run history", "runs")
		if w.runID != "" {
			g.workspaceActions(c, "Delete opened artifact", "delete-artifact")
		}

	case "chat":
		g.projectField(c, "Thread title", &w.thread)
		g.workspaceActions(c, "Browse threads", "threads", "Open thread", "chat")
		for _, thread := range w.choices {
			g.projectButton(c, thread.Label, func() {
				w.thread = thread.ID
				w.choices = nil
				g.workspaceRequest("chat")
			})
		}
		g.projectField(c, "Project message", &w.argument)
		g.workspaceActions(c, "Save pending message", "post", "Message history", "chat")
		if w.runID != "" {
			g.workspaceActions(c, "Save message with last run", "post-with-run")
		}
	}
	return true
}
func (g *Game) workspaceActions(c *widget.Container, labels ...string) {
	w := &g.projects.workspace
	for i := 0; i < len(labels); i += 2 {
		label, action := labels[i], labels[i+1]
		confirm := false
		switch action {
		case "delete-artifact", "delete", "move", "switch", "commit", "discard", "rebase", "resolve", "run", "save":
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
	g.projectField(c, "New name", &w.argument)
	g.workspaceActions(c, "Create empty text file", "create", "Create folder", "mkdir")
}
func (g *Game) workspaceEditor(c *widget.Container, w *workspacePanel) {
	c.AddChild(g.detailText(w.path))
	id, file, line := w.id, w.path, w.line
	g.workspaceLineInput(c, w, func(text string) {
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
	g.workspaceActions(c, "Keep draft", "draft", "Apply draft to file", "save", "Compare checkout text", "compare", "Use current file as base", "rebase", "Discard saved draft", "discard", "Inspect Git conflict", "conflict", "Resolve with saved file", "resolve", "Diff against HEAD", "diff", "Stage file", "stage", "Unstage file", "unstage")
	g.projectField(c, "Commit message or destination path", &w.argument)
	g.workspaceActions(c, "Commit staged file", "commit", "Move file", "move", "Delete file", "delete")
	c.AddChild(g.detailText(strings.Join(w.lines, "\n")))
}

func (g *Game) keepEditingDraft(w *workspacePanel) {
	store, ok := g.runtimeInfo.Queries.(workspaceStore)
	if !ok {
		return
	}
	q := domain.WorkspaceRequest{ProjectID: w.id, Action: "draft", Path: w.path, Base: w.base, Thread: w.thread, Text: strings.Join(w.lines, "\n")}
	_, err := store.Workspace(g.projects.ctx, q)
	if err != nil {
		w.report = "Draft could not be saved: " + err.Error()
		g.layoutDirty = true
	}
}
