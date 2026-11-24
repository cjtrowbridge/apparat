//go:build gui

// Owns local planning-session navigation and unsynchronized turns.
// All state changes run on the HUD thread; session and turn IDs survive navigation.
package gui

import (
	"strings"
	"time"

	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
)

const newPlanningTitle = "New planning session"

type planningSession struct {
	domain.ChatSession
	draft        string
	scroll       float64
	acknowledged bool
}
type planningUI struct {
	modelButton     *widget.Button
	picker          modelPicker
	sessions        []*planningSession
	current         *planningSession
	page, title     string
	pane            *widget.ScrollContainer
	send, saveTitle *widget.Button
	titleField      widget.Focuser
	focusTitle      bool
	voiceSession    *planningSession
}

func (g *Game) newPlanning() {
	p := &g.planning
	if p.current != nil && p.current.Description == newPlanningTitle && len(p.current.Turns) == 0 && g.globalDraft == "" {
		g.openPlanning(p.current)
		return
	}
	now := time.Now().UTC().Truncate(time.Millisecond)
	id, err := domain.NewID(now, nil)
	if err != nil {
		return
	}
	s := &planningSession{ChatSession: domain.ChatSession{ChatHeader: domain.ChatHeader{Version: 1, ID: id, Scope: domain.GlobalChat, Started: now, Description: newPlanningTitle}}, scroll: 1}
	p.sessions = append(p.sessions, s)
	g.openPlanning(s)
}

func (g *Game) detachPlanning() {
	p := &g.planning
	if g.globalField != nil && g.voice.canvas != nil && g.voice.field == g.globalField {
		g.cancelPTT()
	}
	p.voiceSession = nil
	if p.current != nil {
		p.current.draft = g.globalDraft
		if p.pane != nil {
			p.current.scroll = p.pane.ScrollTop
		}
	}
	if g.globalField != nil {
		g.globalField.GetWidget().CustomData.(*documentSurface).ime.Confirm()
		g.ui.ClearFocus()
		g.globalField = nil
	}
	p.pane = nil
	p.send = nil
	p.titleField = nil
	p.saveTitle = nil
}

func (g *Game) openPlanning(s *planningSession) {
	g.detachPlanning()
	g.planning.current = s
	g.globalDraft = s.draft
	g.planning.page = ""
	g.layoutDirty = true
}
func (g *Game) planningPage(page string) {
	g.detachPlanning()
	g.planning.page = page
	if page == "rename" {
		g.planning.title = g.planning.current.Description
		g.planning.focusTitle = true
	}
	g.layoutDirty = true
}
func (g *Game) planningBackAvailable() bool {
	return g.shell.Snapshot().ActiveTab().ID() == hud.TabGlobalSession && g.planning.page != "history"
}
func (g *Game) planningBack() {
	if g.planning.page == "model-picker" {
		g.closeModels()
		return
	}
	if g.planning.page == "rename" {
		g.openPlanning(g.planning.current)
		return
	}
	g.planningPage("history")
}

func (g *Game) enterPlanningVoice() {
	if g.shell.Snapshot().ActiveTab().ID() != hud.TabGlobalSession || g.planning.page != "" {
		g.newPlanning()
	}
	g.openGlobalPrompt()
	g.planning.voiceSession = g.planning.current
}

func (g *Game) submitPlanning() {
	s := g.planning.current
	if s == nil || !g.chatModelReady() || s.Pending() != "" || strings.TrimSpace(g.globalDraft) == "" {
		return
	}
	id, err := domain.NewID(time.Now(), nil)
	if err != nil {
		return
	}
	_, err = s.Apply(id, g.globalDraft, false)
	if err != nil {
		return
	}
	g.globalDraft = ""
	s.draft = ""
	s.scroll = 1
	s.acknowledged = false
	g.planning.pane = nil
	if g.globalField != nil {
		g.ui.ClearFocus()
		g.globalField.GetWidget().CustomData.(*documentSurface).load("")
	}
	g.layoutDirty = true
}

func (g *Game) planningActions(add registerAction, index int, tab hud.Tab) {
	owner := string(tab.ID())
	register := func(id, name string, open func()) {
		add("planning", owner, id, name, tab.Title()+" / "+name, func() {
			g.selectTab(index, nil)
			open()
		})
	}
	register("new", "New planning session", g.newPlanning)
	register("history", "Planning session history", func() {
		g.planningPage("history")
	})
	for _, session := range g.planning.sessions {
		add("planning-session", owner, session.ID, "Resume "+session.Description, tab.Title()+" / "+session.Description+" / "+session.Started.Local().Format("Jan 2 15:04:05"), func() {
			g.selectTab(index, nil)
			g.openPlanning(session)
		})
	}
	if g.planning.current != nil && g.planning.page == "" {
		register("model", "Choose chat model", g.openModels)
		register("rename", "Edit planning session title", func() {
			g.planningPage("rename")
		})
	}
}
func (game *Game) handleDeviceBack(clusterOpen, peerOpen bool) {
	game.workspaceBackEnabled.Store(game.actions.mode != "" || game.inWorkspace() || peerOpen || clusterOpen || game.planningBackAvailable())
	if game.workspaceBackRequested.Swap(false) {
		if game.actions.mode != "" {
			game.backActionScreen()
		} else if game.planningBackAvailable() {
			game.planningBack()
		} else if clusterOpen {
			game.clusterBack()
		} else if peerOpen {
			game.leavePeer()
		} else {
			game.workspaceBack()
		}
	}
}

func (g *Game) updatePlanningControls() {
	p := &g.planning
	if g.shell.Snapshot().ActiveTab().ID() != hud.TabGlobalSession || g.actions.mode != "" {
		return
	}
	if p.send != nil {
		p.send.GetWidget().Disabled = !g.chatModelReady() || strings.TrimSpace(g.globalDraft) == "" || p.current.Pending() != ""
	}
	if p.saveTitle != nil {
		title := strings.TrimSpace(p.title)
		p.saveTitle.GetWidget().Disabled = title == "" || len(title) > 512 || strings.ContainsAny(title, "\r\n")
	}
	if p.focusTitle && p.titleField != nil {
		g.setNavTarget(p.titleField)
		g.ui.SetFocusedWidget(p.titleField)
		p.focusTitle = false
	}
}
