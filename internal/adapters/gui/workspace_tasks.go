//go:build gui

// Presents project tasks, artifacts and local chat in the workspace detail pane.
// Each selected item retains its owning project when navigating back.
package gui

import (
	"encoding/json"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/ebitenui/ebitenui/widget"
	"strconv"
	"strings"
	"time"
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
				g.openProjectConversation(item)
			})
		}
		if len(w.choices) == 0 {
			g.projectButton(c, "General", func() {
				g.openConversation("General")
			})
		}
		if w.cursor != "" {
			g.projectButton(c, "Older sessions", func() {
				g.workspaceRequest("older-sessions")
			})
		}
		g.projectButton(c, "New conversation", g.newProjectChat)
	case "rename-thread":
		g.projectField(c, "Session title", &w.argument)
		g.projectButton(c, "Save title", func() {
			if strings.TrimSpace(w.argument) != "" {
				g.workspaceRequest("agent-rename")
			}
		})
	case "conversation":
		title := w.threadTitle
		bar := controlRow(c)
		bar.AddChild(g.planningIcon("2795", "New conversation", g.newProjectChat), g.detailText(g.planningTitle(title, max(50, g.detailPanePreferredWidth()-170))))
		if w.agentThread && !w.chatDraft {
			bar.AddChild(g.planningIcon("270f", "Edit session title", g.renameProjectChat))
		}
		bubbles := planningColumn()
		g.planning.bubbles = bubbleSelection{}
		for i, message := range w.choices {
			id := message.ID
			if id == "" {
				id = strconv.Itoa(i)
			}
			info := bubbleDetail{job: message.Job, id: w.id + "/" + w.thread + "/" + id}
			_ = json.Unmarshal([]byte(message.Detail), &info.steps)
			g.chatBubble(bubbles, message.Label, !w.agentThread || message.State == "user", message.State == "pending", info)
		}
		if !w.agentThread && len(w.choices) > 0 {
			g.chatBubble(bubbles, "Saved to project · awaiting queue integration", false, false)
		}
		_, t := g.currentChatTerminal()
		if t != nil {
			g.chatViewButtons(c, t)
			w.layout.chatBubbles = bubbles
		} else {
			c.AddChild(bubbles)
		}
		if w.agentThread && !w.chatDraft {
			g.projectButton(bubbles, "Refresh conversation", func() {
				g.workspaceRequest("chat")
			})
		}
		if len(w.choices) == 0 || w.chatError {
			bubbles.AddChild(g.detailText(w.report))
		}

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

func (f *projectForm) documentProject(id string) bool {
	for _, p := range f.rows {
		if p.ID == id {
			return strings.HasPrefix(p.Location, "content://")
		}
	}
	return false
}

func (g *Game) renameProjectChat() {
	if g.projects.pending != nil {
		return
	}
	g.workspaceGo("rename-thread", "")
	g.projects.workspace.argument = g.projects.workspace.threadTitle
}

func (g *Game) newProjectChat() {
	if g.projects.pending != nil {
		return
	}
	w := &g.projects.workspace
	id, err := domain.NewID(time.Now(), nil)
	if err != nil {
		w.report = err.Error()
		g.layoutDirty = true
		return
	}
	w.setThread(id)
	w.threadTitle = "New conversation"
	w.agentThread = !g.projects.documentProject(w.id)
	w.chatDraft = true
	w.chatError = false
	g.workspaceGo("conversation", "")
}
