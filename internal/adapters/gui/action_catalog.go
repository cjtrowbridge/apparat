//go:build gui

// Projects the current HUD and workspace destinations into addressable navigation actions.
// Invocation resolves the address again before entering the existing selection path.
package gui

import (
	"context"
	"ebe-boilerplate/surface"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"slices"
	"strconv"
)

type catalogAction struct {
	hud.Destination
	open func()
}

type registerAction func(kind, owner, id, name, path string, open func())

func (g *Game) actionCatalog() []catalogAction {
	var actions []catalogAction
	add := func(kind, owner, id, name, path string, open func()) {
		key := hud.ActionAddress{Operation: "open", Kind: kind, Owner: owner, Target: id}
		actions = append(actions, catalogAction{hud.Destination{Address: key, Name: name, Path: path}, open})
	}
	for index, tab := range g.shell.Snapshot().Tabs {
		if !tab.Descriptor.Enabled || !tab.Descriptor.Visible {
			continue
		}
		owner := string(tab.ID())
		add("tab", "", owner, tab.Title(), "Tabs / "+tab.Title(), func() {
			g.selectTab(index, nil)
		})
		switch tab.ID() {
		case "stmpck":
			g.registerGameActions(add, index, owner)
		case hud.TabSettings:
			g.registerSettingsActions(add, index, tab)
		default:
			g.registerSectionActions(add, index, tab)
		}
	}
	return actions
}

func (g *Game) registerGameActions(add registerAction, index int, owner string) {
	for id, name := range surface.GameNames() {
		add("game", owner, strconv.Itoa(id), name, "STMPCK / "+name, func() {
			g.selectTab(index, nil)
			session := g.gallerySession()
			if session.Selected == id && !session.Settings {
				return
			}
			for session.Settings || session.Selected >= 0 {
				session.Back()
			}
			session.SetTab(0)
			session.Open(slices.Index(session.Order(), id))
		})
	}
}

func (g *Game) registerSettingsActions(add registerAction, index int, tab hud.Tab) {
	owner := string(tab.ID())
	for _, section := range tab.Sections {
		add("section", owner, section.ActionID, section.Title, "Settings / "+section.Title, func() {
			g.prefs.reveal = section.ActionID
			g.selectTab(index, nil)
		})
	}
	add("screen", owner, "actions", "Command directory", "Settings / Voice Commands / Command directory", g.openActionBrowser)
}

func (g *Game) registerSectionActions(add registerAction, index int, tab hud.Tab) {
	owner := string(tab.ID())
	counts := make(map[string]int)
	if tab.ID() == hud.TabProjects {
		for _, p := range g.projects.rows {
			counts[hud.CommandWords(p.Name)]++
		}
	}
	for i, s := range g.presentedTab(tab).Sections {
		if s.IsSelectorHeading() {
			continue
		}
		kind, key := "section", s.ActionID
		switch {
		case s.ProjectID != "":
			kind, key = "project", s.ProjectID
		case s.ContentKind == hud.ContentChat:
			kind = "chat"
		case projectEntryAction(s.ContentKind):
			key = string(s.ContentKind)
		case s.ContentKind == hud.ContentProject || s.ContentKind == hud.ContentPipeline:
			kind = "project"
		}
		path := tab.Title() + " / " + s.Title
		if s.ProjectID != "" {
			p := g.projects.rows[i-len(projectEntrySections)]
			if counts[hud.CommandWords(p.Name)] > 1 {
				path += " (" + p.Location + ")"
			}
			if p.Available {
				g.registerWorkspaceActions(add, index, i, p, path)
			}
		}
		add(kind, owner, key, s.Title, path, func() {
			g.selectTab(index, nil)
			g.selectSection(tab.ID(), i)
		})
	}
}

func (g *Game) registerWorkspaceActions(add registerAction, index, i int, p domain.Project, path string) {
	selectWorkspace := func() {
		if g.projects.workspace.id != p.ID {
			g.projects.workspace = workspacePanel{id: p.ID, dir: ".", mode: "files"}
		}
		g.selectTab(index, nil)
		g.selectSection(hud.TabProjects, i)
	}
	for _, dest := range workspaceDestinations {
		add("workspace", p.ID, dest.mode, dest.label+" for "+p.Name, path+" / "+dest.label, func() {
			selectWorkspace()
			g.workspaceGo(dest.mode, dest.action)
		})
	}
	threads := p.Threads
	if len(threads) == 0 {
		threads = []domain.WorkspaceChoice{{ID: "General", Label: "General"}}
	}
	w := &g.projects.workspace
	if w.id == p.ID && w.mode == "conversation" && w.thread != "" && !slices.ContainsFunc(threads, func(t domain.WorkspaceChoice) bool {
		return t.ID == w.thread
	}) {
		threads = append(slices.Clone(threads), domain.WorkspaceChoice{ID: w.thread, Label: w.thread})
	}
	for _, thread := range threads {
		add("chat", p.ID, thread.ID, thread.Label, path+" / Chat / "+thread.Label, func() {
			selectWorkspace()
			g.openConversation(thread.ID)
		})
	}
}

func (g *Game) openConversation(id string) {
	g.projects.workspace.setThread(id)
	g.workspaceGo("conversation", "chat")
}

func (g *Game) queryActions(query string, voice bool) hud.ActionMatches {
	actions := g.actionCatalog()
	destinations := make([]hud.Destination, 0, len(actions))
	for _, a := range actions {
		destinations = append(destinations, a.Destination)
	}
	return hud.QueryActions(destinations, query, voice)
}

func (g *Game) invokeAction(id hud.ActionAddress) {
	for _, a := range g.actionCatalog() {
		if a.Address != id {
			continue
		}
		if (id.Owner == string(hud.TabProjects) || id.Kind == "workspace" || id.Kind == "chat" && id.Owner != string(hud.TabComrades)) && g.projects.pending != nil {
			g.actionNotice("Projects is busy. Try again when the current operation finishes.", nil)
			return
		}
		g.closeActionScreen()
		a.open()
		return
	}
	g.actionNotice("That destination is no longer available.", nil)
}

// During the initial registry read, an apparent unique name may still collide.
func (g *Game) awaitProjectCatalog() bool {
	f := &g.projects
	if f.store == nil || !f.catalogInitialized || f.catalogLoaded {
		return false
	}
	if f.pending == nil {
		g.changeProjects(func(context.Context) error {
			return nil
		})
	}
	return true
}
