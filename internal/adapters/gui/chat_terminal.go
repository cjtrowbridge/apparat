//go:build gui

// Associates terminal selection, drafts and asynchronous owner requests with one chat.
// Terminal operations keep their own request state so they never replace the chat composer.
package gui

import (
	"context"
	et "github.com/hajimehoshi/ebiten/v2/text/v2"
	"image/color"
	"strings"
	"time"

	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/ebitenui/ebitenui/widget"
)

type chatTerminalStore interface {
	ChatTerminal(context.Context, domain.WorkspaceRequest) (domain.WorkspaceView, error)
}

type chatTerminalPane struct {
	controls                                             bool
	history, historyCursor                               string
	readingHistory                                       bool
	hosts                                                bool
	host                                                 string
	split                                                int
	view, selected, draft, screen, state, notice, cursor string
	permission, handled, tracking                        string
	permissionPage                                       bool
	chooser, initialized                                 bool
	choices                                              []domain.WorkspaceChoice
	pending                                              chan workspaceReply
	next                                                 time.Time
}

func (g *Game) currentChatTerminal() (modelAddress, *chatTerminalPane) {
	a, ok := g.modelAddress()
	if !ok || !domain.ValidID(a.session) || g.inWorkspace() && !g.projects.workspace.agentThread {
		return a, nil
	}
	if g.planning.terminals == nil {
		g.planning.terminals = make(map[modelAddress]*chatTerminalPane)
	}
	t := g.planning.terminals[a]
	if t == nil {
		t = &chatTerminalPane{}
		g.planning.terminals[a] = t
	}
	return a, t
}

func (g *Game) requestChatTerminal(a modelAddress, t *chatTerminalPane, action, id, arg string) {
	if t.pending != nil {
		t.notice = "Terminal busy. Try again when the request finishes."
		g.layoutDirty = true
		return
	}
	store, ok := g.runtimeInfo.Queries.(chatTerminalStore)
	if !ok {
		t.notice = "Terminal host unavailable."
		return
	}
	q := domain.WorkspaceRequest{Action: "chat-terminal-" + action, Thread: a.session, ProjectID: a.project, ID: id, Argument: arg}
	wait := g.currentTerminalWait(a)
	if wait != nil {
		q.Correlation = wait.Job
	}
	if wait != nil && (action == "approved" || action == "result") {
		q.Approved = action == "approved" && t.permission == "always-ask"
		q.Deadline = wait.Deadline
	}
	result := make(chan workspaceReply, 1)
	t.pending = result
	parent := g.runContext
	if parent == nil {
		parent = context.Background()
	}
	g.projects.workers.Go(func() {
		ctx, cancel := context.WithTimeout(parent, 10*time.Second)
		defer cancel()
		v, err := store.ChatTerminal(ctx, q)
		result <- workspaceReply{q, v, err}
	})
}

func (g *Game) pollChatTerminal() {
	a, t := g.currentChatTerminal()
	if t == nil {
		return
	}
	if a.project == "" && !g.ownsPlanningTerminal() {
		return
	}
	wait := g.currentTerminalWait(a)
	if t.pending != nil {
		select {
		case r := <-t.pending:
			t.pending = nil
			t.next = time.Now().Add(time.Second)
			if g.applyChatTerminalCallback(a, t, wait, r) {
				return
			}
			if r.err != nil {
				t.notice = r.err.Error()
				g.layoutDirty = true
				return
			}
			v := r.view
			if r.request.Action == "chat-terminal-history" {
				t.history, t.historyCursor, t.readingHistory = v.Report, v.Cursor, true
				g.layoutDirty = true
				return
			}
			if r.request.Action == "chat-terminal-hosts" {
				t.choices, t.notice, t.hosts = v.Choices, v.Report, true
				g.layoutDirty = true
				return
			}
			if r.request.Action == "chat-terminal-bind" {
				t.host, t.hosts, t.initialized, t.chooser, t.notice = v.Detail, false, false, false, ""
				g.layoutDirty = true
				return
			}
			if v.Base != "" {
				t.host = v.Base
			}
			if r.request.Action == "chat-terminal-list" {
				if r.request.Argument == "" {
					t.choices = nil
				}
				t.choices = append(t.choices, v.Choices...)
				t.cursor = v.Cursor
				t.notice = v.Report
				g.layoutDirty = true
				return
			}
			if t.screen != v.Report || t.state != v.Detail || t.notice != "" || t.selected != r.request.ID {
				t.screen, t.state, t.notice, t.selected = v.Report, v.Detail, "", r.request.ID
				g.layoutDirty = true
			}
			if r.request.Action == "chat-terminal-start" {
				t.chooser = false
				g.layoutDirty = true
			}
		default:
		}
		return
	}
	if t.permission == "" && (t.view != "" || t.permissionPage || wait != nil) {
		g.requestChatTerminal(a, t, "permissions", "", "")
		return
	}
	if g.serviceChatTerminalWait(a, t, wait) || t.view == "" {
		return
	}
	if !t.initialized {
		t.initialized = true
		t.selected = a.session
		g.requestChatTerminal(a, t, "start", t.selected, "")
		return
	}
	if !t.chooser && !t.readingHistory && t.notice == "" && t.state == "Running" && time.Now().After(t.next) {
		g.requestChatTerminal(a, t, "screen", t.selected, "")
	}
}

func (g *Game) chatViewButtons(c *widget.Container, t *chatTerminalPane) {
	row := controlRow(c)
	for _, option := range []struct{ icon, label, view string }{
		{"1f4ac", "Chat", ""}, {"2194", "Split chat and terminal", "split"}, {"1f5a5", "Terminal", "terminal"},
	} {
		b := g.planningIcon(option.icon, option.label, func() {
			t.view = option.view
			a, ok := g.modelAddress()
			if ok && a.project == "" && option.view != "" {
				g.planning.current.scroll = 0
				g.planning.pane = nil
			}
			g.layoutDirty = true
		})
		row.AddChild(b)
	}
}

func (g *Game) chatTerminalContent(c *widget.Container, a modelAddress, t *chatTerminalPane) {
	if a.project == "" && !g.ownsPlanningTerminal() {
		message := "Open this terminal on the device where its global chat was created."
		if g.planning.current.TerminalHost == "" {
			message = "Send a message from this device to assign the terminal for this existing chat."
		}
		c.AddChild(g.detailText(message))
		return
	}
	host := a.project
	if host == "" {
		host = t.host
	}
	if host == "" {
		host = "This device · local clone directory"
	}
	c.AddChild(g.detailText(host))
	toolbar := controlRow(c)
	toolbar.AddChild(g.planningIcon("1f519", "Terminal sessions", func() {
		t.chooser, t.readingHistory, t.hosts = true, false, false
		g.requestChatTerminal(a, t, "list", "", "")
		g.layoutDirty = true
	}), g.planningIcon("2795", "New terminal", func() {
		id, err := domain.NewID(time.Now(), nil)
		if err == nil {
			g.requestChatTerminal(a, t, "start", id, "")
		}
	}))
	if t.notice != "" {
		c.AddChild(g.detailText(t.notice))
	}
	if t.hosts {
		for _, p := range t.choices {
			g.projectButton(c, p.Label, func() {
				g.requestChatTerminal(a, t, "bind", "", p.ID)
			})
		}
		return
	}
	if a.project == "" && t.host == "" && t.notice != "" {
		g.projectButton(c, "Choose worker terminal host", func() {
			g.requestChatTerminal(a, t, "hosts", "", "")
		})
	}
	if t.chooser {
		for _, item := range t.choices {
			g.projectButton(c, item.Label, func() {
				t.chooser = false
				t.selected, t.screen, t.state = item.ID, "", ""
				g.requestChatTerminal(a, t, "screen", item.ID, "")
				g.layoutDirty = true
			})
		}
		if t.cursor != "" {
			g.projectButton(c, "More sessions", func() {
				g.requestChatTerminal(a, t, "list", "", t.cursor)
			})
		}
		return
	}
	if t.readingHistory {
		c.AddChild(g.detailText(t.history))
		g.projectButton(c, "Return to live screen", func() {
			t.readingHistory = false
			g.requestChatTerminal(a, t, "screen", t.selected, "")
			g.layoutDirty = true
		})
		if t.historyCursor != "" {
			g.projectButton(c, "Older output", func() {
				g.requestChatTerminal(a, t, "history", t.selected, t.historyCursor)
			})
		}
		return
	}
	if t.controls {
		g.projectButton(c, "Back to terminal", func() {
			t.controls = false
			g.layoutDirty = true
		})
		g.chatTerminalKeys(c, a, t)
		return
	}
	if t.screen != "" {
		var face et.Face = &et.GoTextFace{Source: terminalFont, Size: float64(max(10, min(16, g.prefs.Font)))}
		c.AddChild(widget.NewText(widget.TextOpts.Text(strings.TrimRight(t.screen, " \n"), &face, color.White)))

	}
	if t.state != "Running" {
		c.AddChild(g.detailText(t.state))
		return
	}
	field := g.multilineInput(c, "Terminal input", &t.draft)
	field.submit = func() {
		g.requestChatTerminal(a, t, "input", t.selected, t.draft+"\r")
	}
	g.projectButton(c, "Send line", field.submit)
	g.projectButton(c, "Terminal controls", func() {
		t.controls = true
		g.layoutDirty = true
	})
}
func (g *Game) chatTerminalKeys(c *widget.Container, a modelAddress, t *chatTerminalPane) {
	g.projectButton(c, "Retained output", func() {
		t.controls = false
		g.requestChatTerminal(a, t, "history", t.selected, "")
	})
	for _, key := range []struct{ name, bytes string }{{"Tab", "\t"}, {"Escape", "\x1b"}, {"Interrupt", "\x03"}, {"Up", "\x1b[A"}, {"Down", "\x1b[B"}, {"Left", "\x1b[D"}, {"Right", "\x1b[C"}} {
		g.projectButton(c, key.name, func() {
			g.requestChatTerminal(a, t, "input", t.selected, key.bytes)
		})
	}
	for _, size := range []string{"40x20", "80x24", "120x40"} {
		g.projectButton(c, size, func() {
			g.requestChatTerminal(a, t, "resize", t.selected, size)
		})
	}
	g.projectButton(c, "Take control", func() {
		g.requestChatTerminal(a, t, "take", t.selected, "")
	})
	g.projectButton(c, "End terminal", func() {
		g.requestChatTerminal(a, t, "stop", t.selected, "")
	})
}

func (g *Game) chatPanes(parent *widget.Container, bubbles *widget.Container, a modelAddress, t *chatTerminalPane) {
	g.approvalControls(parent, a, t)
	if t.view == "" {
		parent.AddChild(bubbles)
		return
	}
	term := planningColumn()
	g.chatTerminalContent(term, a, t)
	if t.view == "terminal" {
		parent.AddChild(term)
		return
	}
	width := g.chatSplitAvailable(a)
	left := g.chatSplitWidth(width, t)
	bubbles.GetWidget().MinWidth = left
	split := widget.NewContainer(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})), widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(3), widget.GridLayoutOpts.Stretch([]bool{false, false, true}, []bool{true}), widget.GridLayoutOpts.Spacing(2, 0))))
	split.AddChild(boundPreferredWidth(bubbles, func() int {
		return left
	}), g.chatDivider(), boundPreferredWidth(term, func() int {
		return max(44, width-left-48)
	}))
	parent.AddChild(split)
}
