//go:build gui

// Maps each Projects destination to its content and its relevant actions.
// The Back action follows the same hierarchy on touch, Android and controllers.
package gui

import (
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
	"path"
	"strings"
)

func (g *Game) inWorkspace() bool {
	if g.shell.Snapshot().ActiveTab().ID() != hud.TabProjects {
		return false
	}
	i := g.selectedSections[hud.TabProjects] - 2
	return i >= 0 && i < len(g.projects.rows) && g.projects.rows[i].ID == g.projects.workspace.id && (!g.collapsed() || g.detailOpen[hud.TabProjects])
}
func (g *Game) RequestWorkspaceBack() {
	g.workspaceBackRequested.Store(true)
}
func (g *Game) WorkspaceBackAvailable() bool {
	return g.workspaceBackEnabled.Load()
}
func (g *Game) workspaceGo(mode, action string) {
	if g.projects.pending != nil {
		return
	}
	w := &g.projects.workspace
	g.stopFolderCheck()
	if mode == "changes" {
		w.changes = nil
	}
	w.mode = mode
	w.report = ""
	w.confirmation = ""
	w.argument = ""
	w.choices = nil
	w.autoEdit = false
	g.layoutDirty = true
	if action != "" {
		g.workspaceRequest(action)
	}
}
func (g *Game) workspaceBack() {
	if g.projects.pending != nil {
		return
	}
	w := &g.projects.workspace
	switch w.mode {
	case "files":
		if w.dir != "." {
			w.dir = path.Dir(w.dir)
			g.workspaceGo("files", "list")
			return
		}
		g.stopFolderCheck()
		*w = workspacePanel{}
		delete(g.selectedSections, hud.TabProjects)
		g.showList(hud.TabProjects)
	case "review", "history", "branches", "transactions":
		g.workspaceGo("changes", "changes")
	case "commit", "conflict":
		g.workspaceGo("review", "review")
	case "branch":
		g.workspaceGo("branches", "branches")
	case "commit-detail":
		g.workspaceGo("history", "history")
	case "task", "runs":
		g.workspaceGo("tasks", "tasks")
	case "result":
		g.workspaceGo("runs", "runs")
	case "output", "output-options":
		g.workspaceGo("result", "")
	case "conversation":
		g.workspaceGo("chat", "threads")
	case "attachments":
		g.workspaceGo("conversation", "chat")
	case "move", "file-options", "inline":
		g.workspaceGo("file", "read")
	default:
		g.workspaceGo("files", "list")
	}
}
func (g *Game) workspaceScreen(c *widget.Container, w *workspacePanel) {
	btn := func(label string, fn func()) {
		g.projectButton(c, label, func() {
			if g.projects.pending == nil {
				fn()
			}
		})
	}
	goTo := func(label, mode, action string) {
		btn(label, func() {
			g.workspaceGo(mode, action)
		})
	}
	nav := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionHorizontal))))
	nav.AddChild(widget.NewButton(
		widget.ButtonOpts.Graphic(&widget.GraphicImage{Idle: emojiIcon("1f519")}),
		widget.ButtonOpts.GraphicPadding(widget.Insets{}),
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(54, 48)),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			if !g.bodySelectionSuppressed() {
				g.workspaceBack()
			}
		}),
	))
	roots := w.mode == "files" || w.mode == "changes" || w.mode == "tasks" || w.mode == "chat"
	if roots {
		for _, dest := range []struct{ label, mode, action string }{{"Files", "files", "list"}, {"Changes", "changes", "changes"}, {"Tasks", "tasks", "tasks"}, {"Chat", "chat", "threads"}} {
			g.workspaceNavButton(nav, dest.label, w.mode == dest.mode, func() {
				g.workspaceGo(dest.mode, dest.action)
			})
		}
	}
	c.AddChild(nav)
	if g.projects.pending != nil {
		c.AddChild(g.detailText("Loading " + w.mode + "…"))
	}
	switch w.mode {
	case "files":
		g.workspaceBrowser(c, w)
		goTo("Folder options", "folder-options", "")
	case "folder-options":
		goTo("New file", "create", "")
		goTo("New folder", "mkdir", "")
		goTo("Repository details", "registration", "")
	case "create", "mkdir", "move":
		label := "Name"
		if w.mode == "move" {
			label = "Destination path"
		}
		g.projectField(c, label, &w.argument)
		btn("Apply", func() {
			if w.mode == "create" {
				w.autoEdit = true
			}
			g.workspaceRequest(w.mode)
		})
	case "file":
		c.AddChild(g.detailText(w.path))
		btn("Edit document", g.editDocument)
		goTo("Changes", "changes", "changes")
		goTo("File options", "file-options", "")
		c.AddChild(g.detailText(strings.Join(w.lines, "\n")))
	case "file-options":
		goTo("Rename or move", "move", "")
		g.workspaceActions(c, "Delete "+w.path, "delete", "Discard draft", "discard")
		goTo("Line editor fallback", "inline", "")
	case "inline":
		g.workspaceEditor(c, w)
	case "changes", "review", "commit", "history", "branches", "branch", "commit-detail", "transactions", "conflict":
		g.changesScreen(c, w)
		return
	default:
		g.taskChatScreen(c, w)
		return
	}
	if w.report != "" {
		c.AddChild(g.detailText(w.report))
	}
}
func (g *Game) changesScreen(c *widget.Container, w *workspacePanel) {
	goTo := func(label, mode, action string) {
		g.projectButton(c, label, func() {
			g.workspaceGo(mode, action)
		})
	}
	switch w.mode {
	case "changes":
		if len(w.changes) == 0 && w.report == "" && g.projects.pending == nil {
			c.AddChild(g.detailText("Clean working tree"))
		}
		for _, item := range w.changes {
			state := "Unstaged"
			if item.Staged {
				state = "Staged"
			}
			if item.Staged && item.Unstaged {
				state = "Staged + unstaged"
			}
			if item.Conflict {
				state = "Conflict"
			}
			g.workspaceRow(c, item.Path+" · "+state, func() {
				if g.projects.pending != nil {
					return
				}
				w.path = item.Path
				w.change = item
				w.base = ""
				g.workspaceGo("review", "review")
			})
		}
		goTo("History", "history", "history")
		goTo("Branches", "branches", "branches")
		goTo("Save history", "transactions", "transactions")
	case "review":
		c.AddChild(g.detailText(w.path))
		if w.base != "" {
			stage := "stage"
			if w.base == "deleted" {
				stage = "stage-deletion"
			}
			if w.change.Unstaged || !w.change.Staged {
				g.workspaceActions(c, "Stage file", stage)
			}
			if w.change.Staged {
				g.workspaceActions(c, "Unstage file", "unstage")
				goTo("Commit…", "commit", "")
			}
			if w.change.Conflict {
				goTo("Review conflict", "conflict", "conflict")
			}
		}
	case "commit":
		c.AddChild(g.detailText("Commit " + w.path + " to " + g.workspaceBranch()))
		g.projectField(c, "Commit message", &w.argument)
		g.projectButton(c, "Commit file", func() {
			g.workspaceRequest("commit")
		})
	case "conflict":
		g.workspaceActions(c, "Resolve using saved file", "resolve")
	case "history", "branches":
		for line := range strings.SplitSeq(strings.TrimSpace(w.report), "\n") {
			if line == "" {
				continue
			}
			if w.mode == "branches" {
				g.projectButton(c, line, func() {
					g.workspaceGo("branch", "")
					w.argument = line
				})
				continue
			}
			id, title, ok := strings.Cut(line, " ")
			if !ok || len(id) != 40 {
				c.AddChild(g.detailText(line))
				continue
			}
			g.projectButton(c, title+" · "+id[:8], func() {
				g.workspaceGo("commit-detail", "")
				w.argument = id
				g.workspaceRequest("commit-detail")
			})
		}
		return
	case "branch":
		c.AddChild(g.detailText(w.argument))
		g.workspaceActions(c, "Switch branch", "switch")
	}
	if w.report != "" && w.mode == "review" {
		g.diffText(c, w.report)
	} else if w.report != "" {
		c.AddChild(g.detailText(w.report))
	}
}
func (g *Game) workspaceBranch() string {
	for _, p := range g.projects.rows {
		if p.ID == g.projects.workspace.id {
			return p.Branch
		}
	}
	return ""
}
func (g *Game) taskChatScreen(c *widget.Container, w *workspacePanel) {
	goTo := func(label, mode, action string) {
		g.projectButton(c, label, func() {
			g.workspaceGo(mode, action)
		})
	}
	switch w.mode {
	case "tasks", "runs", "attachments":
		for _, item := range w.choices {
			g.workspaceRow(c, item.Label, func() {
				if g.projects.pending != nil {
					return
				}
				switch w.mode {
				case "attachments":
					w.attachment = item.ID
					w.attachmentLabel = item.Label
					g.workspaceGo("conversation", "chat")
				case "tasks":
					w.selected = item
					g.workspaceGo("task", "")
				case "runs":
					w.selected = item
					w.runID = item.ID
					g.workspaceGo("result", "")
				}
			})
		}
		if w.mode == "tasks" {
			goTo("Run history", "runs", "runs")
		}
	case "task":
		c.AddChild(g.detailText(w.selected.Label))
		c.AddChild(g.detailText(w.selected.Detail))
		g.projectButton(c, "Run", func() {
			w.argument = w.selected.ID
			g.workspaceRequest("run")
		})
	case "result":
		c.AddChild(g.detailText(w.selected.Label))
		if w.selected.Detail != "" {
			c.AddChild(g.detailText(w.selected.Detail))
		}
		if w.selected.State == "completed" || w.selected.State == "" {
			g.projectButton(c, "Open output", func() {
				w.argument = w.runID
				g.workspaceRequest("artifact")
			})
		}
	case "output":
		c.AddChild(g.detailText(w.report))
		goTo("Output details", "output-options", "")
		return
	case "output-options":
		c.AddChild(g.detailText(w.outputDetails))
		g.workspaceActions(c, "Delete output", "delete-artifact")
	case "chat":
		for _, item := range w.choices {
			g.workspaceRow(c, item.Label, func() {
				w.setThread(item.ID)
				g.workspaceGo("conversation", "chat")
			})
		}
		if len(w.choices) == 0 {
			g.projectButton(c, "General", func() {
				w.setThread("General")
				g.workspaceGo("conversation", "chat")
			})
		}
		goTo("New conversation", "new-thread", "")
	case "new-thread":
		g.projectField(c, "Conversation name", &w.argument)
		g.projectButton(c, "Open", func() {
			name := strings.TrimSpace(w.argument)
			if name != "" {
				w.setThread(name)
				g.workspaceGo("conversation", "chat")
			}
		})
	case "conversation":
		c.AddChild(g.detailText(w.thread))
		c.AddChild(g.detailText(w.report))

		return
	}
	if w.report != "" {
		c.AddChild(g.detailText(w.report))
	}
}

func (w *workspacePanel) setThread(name string) {
	if w.thread != name {
		w.message = ""
		w.attachment = ""
		w.attachmentLabel = ""
	}
	w.thread = name
}
