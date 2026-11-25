//go:build gui

// Shows the exact pending terminal action, sends a single human decision and polls for exit.
// The transcript owner fences stale callbacks and resumes its model loop independently.
package gui

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/ebitenui/ebitenui/widget"
)

func (g *Game) ownsPlanningTerminal() bool {
	store, ok := g.runtimeInfo.Queries.(interface{ ChatTerminalDevice() string })
	s := g.planning.current
	return ok && s != nil && s.TerminalHost != "" && store.ChatTerminalDevice() == s.TerminalHost
}

func (g *Game) currentTerminalWait(a modelAddress) *domain.ChatTerminalWait {
	if a.project == "" {
		return g.planning.current.TerminalWait()
	}
	return g.projects.workspace.terminalWait
}

func (g *Game) terminalResult(a modelAddress, t *chatTerminalPane, job, content string, failed bool) {
	if len(content) > 4096 {
		content = strings.ToValidUTF8(content[:4088], "") + "…"
	}
	if !utf8.ValidString(content) {
		content = strings.ToValidUTF8(content, "�")
	}
	if a.project == "" {
		base := "ok"
		if failed {
			base = "failed"
		}
		if g.savePlanningOperation("global-terminal-result", domain.GlobalRequest{Session: a.session, Turn: job, Input: content, Base: base}) {
			t.handled = job
		}
		return
	}
	store, ok := g.runtimeInfo.Queries.(workspaceStore)
	if !ok {
		return
	}
	q := domain.WorkspaceRequest{ProjectID: a.project, Path: ".", Thread: a.session, ID: job, Action: "agent-terminal-result", Approved: !failed, Argument: content}
	ch := make(chan workspaceReply, 1)
	t.pending = ch
	g.projects.workers.Go(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		v, e := store.Workspace(ctx, q)
		ch <- workspaceReply{q, v, e}
	})
}

func (g *Game) applyChatTerminalCallback(a modelAddress, t *chatTerminalPane, w *domain.ChatTerminalWait, r workspaceReply) bool {
	switch r.request.Action {
	case "chat-terminal-permissions", "chat-terminal-permission-set":
		if r.err != nil {
			t.notice = r.err.Error()
			t.permission = "unavailable"
		} else {
			t.permission = r.view.Detail
			t.host = r.view.Base
			t.notice = ""
		}
	case "agent-terminal-result":
		if r.err != nil {
			t.notice = r.err.Error()
		} else {
			t.handled = r.request.ID
		}
	case "chat-terminal-approved", "chat-terminal-result", "chat-terminal-observe":
		if w == nil || w.Job != r.request.Correlation && w.Job != r.request.ID {
			return true
		}
		if r.err != nil {
			g.terminalResult(a, t, w.Job, r.err.Error(), true)
		} else if w.Observe || r.view.Detail != "Running" {
			g.terminalResult(a, t, w.Job, r.view.Detail+"\n"+r.view.Report, !w.Observe && r.view.Detail != "Exited 0")
		} else {
			t.tracking = w.Job
		}
	default:
		return false
	}
	g.layoutDirty = true
	return true
}

func (g *Game) serviceChatTerminalWait(a modelAddress, t *chatTerminalPane, w *domain.ChatTerminalWait) bool {
	if w == nil || t.handled == w.Job || !time.Now().Before(w.Deadline) {
		return false
	}
	if w.Observe {
		id := t.selected
		if id == "" {
			id = a.session
		}
		g.requestChatTerminal(a, t, "observe", id, "")
		return true
	}
	if t.tracking == w.Job {
		return false
	}
	if t.permission == "yolo" {
		g.requestChatTerminal(a, t, "approved", w.Job, w.Command)
		return true
	}
	if t.permission != "always-ask" {
		g.terminalResult(a, t, w.Job, "Read-Only. The command did not run.", true)
		return true
	}
	return false
}

func (g *Game) chatPermissionButton(row *widget.Container, a modelAddress, t *chatTerminalPane) {
	if t == nil {
		return
	}
	icon := "1f512"
	if t.permission == "always-ask" {
		icon = "1f50d"
	} else if t.permission == "yolo" {
		icon = "1f985"
	}
	b := g.planningIcon(icon, "Chat permissions", func() {
		t.permissionPage = true
		g.layoutDirty = true
	})
	row.AddChild(b)
}

func (g *Game) chatPermissionPage(c *widget.Container, a modelAddress, t *chatTerminalPane) {
	c.AddChild(g.planningIcon("1f519", "Back to chat", func() {
		t.permissionPage = false
		g.layoutDirty = true
	}), g.detailText("Chat permissions"))
	for _, o := range []struct{ label, icon, mode, help string }{
		{"Read-Only", "1f512", "read-only", "You control the terminal. The agent can observe its screen."},
		{"Always Ask", "1f50d", "always-ask", "Review and approve each agent command or other tool that makes changes."},
		{"Yolo", "1f985", "yolo", "Run commands automatically in this chat’s isolated terminal. It can read the mounted checkout and write temporary files. Network access is disabled."},
	} {
		row := controlRow(c)
		b := g.planningIcon(o.icon, o.label, func() {
			g.requestChatTerminal(a, t, "permission-set", "", o.mode)
			t.permissionPage = false
			g.layoutDirty = true
		})
		b.GetWidget().Disabled = o.mode == "" || t.pending != nil || a.project == "" && !g.ownsPlanningTerminal()
		row.AddChild(b, g.detailText(o.label))
		c.AddChild(g.detailText(o.help))
	}
}

func (g *Game) approvalControls(c *widget.Container, a modelAddress, t *chatTerminalPane) {
	w := g.currentTerminalWait(a)
	if w == nil || w.Observe || t.handled == w.Job || t.permission != "always-ask" {
		return
	}
	host := a.project
	if host == "" {
		host = t.host
	}
	if host == "" {
		host = "This device · local clone directory"
	}
	c.AddChild(g.detailText("Requested command · "+host), g.detailText(fmt.Sprintf("%q", w.Command)))
	if t.tracking == w.Job {
		c.AddChild(g.detailText("Waiting for command exit"))
		return
	}
	g.projectButton(c, "Approve once", func() {
		if t.pending == nil && g.currentTerminalWait(a) == w {
			g.requestChatTerminal(a, t, "approved", w.Job, w.Command)
		}
	})
	g.projectButton(c, "Deny", func() {
		if t.pending == nil && g.currentTerminalWait(a) == w {
			g.terminalResult(a, t, w.Job, "Denied by user.", true)
		}
	})
}
