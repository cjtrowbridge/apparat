//go:build gui

// Exercises terminal clipboard operations through the rendered chat controls.
// Verifies literal bytes, frozen display, view changes and stale session callbacks.
package gui

import (
	"github.com/cjtrowbridge/apparat/internal/domain"
	"testing"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/event"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestTerminalOutputClipboard(t *testing.T) {
	for _, width := range []int{390, 1280} {
		g := horizontalGame()
		g.Layout(width, 844)
		selectTabThroughStrip(t, g, hud.TabGlobalSession)
		attachPlanningTestStore(g)
		g.planning.current.TerminalHost = "test-device"
		_, pane := g.currentChatTerminal()
		pane.selected = "session-a"
		pane.state = "Running"
		pane.view = "terminal"
		value := "  café\tλ\n  tail"
		pane.screen = value
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		o := g.terminalOutput
		p := &o.passage
		origin := p.view.Rect.Min
		if !g.beginBubble(origin.X+5, origin.Y+5, false, 0) {
			t.Fatal("live terminal did not receive mouse selection")
		}
		g.selectPassage(0, len(value))
		g.finishBubble()
		o.live.enqueue("")
		pane.screen = "new revision"
		renderNavigation(g)
		if p.selected() != value || p.copy.GetWidget().Disabled {
			t.Fatal("lost selection on refresh")
		}
		var copied string
		g.runtimeInfo.CopyText = func(s string) error {
			copied = s
			return nil
		}
		g.bodyScrollCancelUpdates = 0
		p.copy.Click()
		event.ExecuteDeferred()
		if copied != value {
			t.Fatal("copy changed terminal bytes", copied)
		}
		pane.view = "split"
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		if g.terminalOutput != o || p.selected() != value {
			t.Fatal("same terminal lost snapshot")
		}
		g.runtimeInfo.PasteText = func() (string, error) {
			return "echo café", nil
		}
		cPressed := func(key ebiten.Key) bool {
			return key == ebiten.KeyC
		}
		copied = ""
		o.live.controlKeys(true, cPressed)
		if copied != value || pane.queue != "" {
			t.Fatal("copy chord sent shell input")
		}
		o.live.controlKeys(false, cPressed)
		if pane.queue != "\x03" {
			t.Fatal("interrupt chord did not send ETX")
		}
		pane.queue = ""
		o.live.controlKeys(true, func(key ebiten.Key) bool {
			return key == ebiten.KeyV
		})
		if pane.queue != "echo café" || p.selected() != "" {
			t.Fatal("paste added bytes or kept frozen output")
		}
		stale := o.live
		pane.selected = "session-b"
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		stale.paste()
		if pane.queue != "echo café" || g.terminalOutput == o {
			t.Fatal("stale paste reached new terminal")
		}
		pane.history = value
		pane.readingHistory = true
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		h := g.terminalOutput
		if h.live != nil || h.source != value || h.id.page != "history" {
			t.Fatal("history source")
		}
		g.bubbleText.pending = &h.passage
		g.selectPassage(2, len("  café"))
		h.copySelection()
		if copied != "café" {
			t.Fatal("history copy")
		}
		pane.readingHistory = false
		pane.view = ""
		g.rebuildUI(g.shell.Snapshot())
		if g.terminalOutput != nil || len(g.terminalClipboardActions()) != 0 {
			t.Fatal("stale terminal actions")
		}
		g.Close()
	}
}

func TestTerminalLiteralWrap(t *testing.T) {
	value := "\t git **café**\n\n    two repeated repeated\n"
	for _, width := range []int{44, 400} {
		m := literalTerminalLayout(value, width, 14, true)
		if m.plain.String() != value {
			t.Fatal("literal output changed", m.plain.String())
		}
		for _, r := range m.runs {
			if value[r.start:r.start+len(r.value)] != r.value {
				t.Fatal("bad hit-test range")
			}
		}
	}
}

func TestTerminalHitTestPreservesRows(t *testing.T) {
	m := literalTerminalLayout("long first row\nx\n\n\tend", 400, 14, false)
	line := float64(terminalLineHeight(14))
	for _, c := range []struct {
		x, y float64
		want int
	}{{300, line + 1, 16}, {300, 2*line + 1, 17}, {0, 3*line + 1, 18}, {50, 3*line + 1, 19}} {
		got := m.position(c.x, c.y)
		if got != c.want {
			t.Errorf("position(%v,%v)=%d, want %d", c.x, c.y, got, c.want)
		}
	}
}

func TestProjectTerminalClipboardViews(t *testing.T) {
	g := horizontalGame()
	defer g.Close()
	g.projects.store = layoutRegistry{}
	g.projects.rows = []domain.Project{{ID: "p", Name: "workspace", Available: true}}
	selectTabThroughStrip(t, g, hud.TabProjects)
	w := &g.projects.workspace
	*w = workspacePanel{id: "p", thread: "01M7QZQGAVAVCRGVB7ET1A5QGV", agentThread: true, mode: "conversation"}
	g.detailOpen[hud.TabProjects] = true
	g.selectSection(hud.TabProjects, len(projectEntrySections))
	_, pane := g.currentChatTerminal()
	pane.view, pane.state, pane.selected, pane.screen = "split", "Running", "s", "clipboard"
	for _, page := range []string{"split", "terminal", "ended", "history", "terminal-session", "terminal-history"} {
		switch page {
		case "split", "terminal":
			pane.view = page
		case "ended":
			pane.state = "Ended"
		case "history":
			pane.readingHistory = true
			pane.history = "clipboard"
		default:
			w.mode, w.runID, w.report = page, "standalone", "clipboard"
		}
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		o := g.terminalOutput
		if o == nil || o.source != "clipboard" || o.id.chat.project != "p" {
			t.Fatal("missing project terminal", page)
		}
		origin := o.passage.view.Rect.Min
		if !g.beginBubble(origin.X+5, origin.Y+5, true, 0) {
			t.Fatal("touch did not reach terminal", page, origin)
		}
		g.selectPassage(0, 4)
		g.finishBubble()
		var copied string
		g.runtimeInfo.CopyText = func(value string) error {
			copied = value
			return nil
		}
		o.copySelection()
		if copied != "clip" {
			t.Fatal("project clipboard", page)
		}
	}
}
