//go:build gui

// Presents project tasks, artifacts and local chat in the workspace detail pane.
// Each selected item retains its owning project when navigating back.
package gui

import (
	"github.com/ebitenui/ebitenui/widget"
	"strings"
)

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
		if w.mode != "tasks" && w.cursor != "" {
			g.projectButton(c, "Older runs", func() {
				w.argument = w.cursor
				g.workspaceRequest("runs")
			})
		}
		if w.mode == "tasks" {
			if w.cursor != "" {
				g.projectButton(c, "Next task page", func() {
					w.argument = w.cursor
					g.workspaceRequest("tasks")
				})
			}
			goTo("Run history", "runs", "runs")
		}
	case "task":
		c.AddChild(g.detailText(w.selected.Label))
		c.AddChild(g.detailText(w.selected.Detail))
		if w.selected.State == "definition" {
			return
		}
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
				g.openConversation(item.ID)
			})
		}
		if len(w.choices) == 0 {
			g.projectButton(c, "General", func() {
				g.openConversation("General")
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
