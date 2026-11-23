//go:build gui

// Keeps one registry page per host and exposes continuations through the selector.
// Replacing a page clears the selection so an old index cannot open a different checkout.
package gui

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"os"
	"strings"
)

type projectPager interface {
	ProjectPage(context.Context, string, string) ([]domain.Project, error)
}
type projectTurn struct{ host, cursor, label string }

func projectOwner(p domain.Project) string {
	host, _, ok := strings.Cut(p.ID, "/")
	if ok && p.RemoteHost != "" {
		return host
	}
	return "local"
}
func (f *projectForm) pageTurns() []projectTurn {
	var turns []projectTurn
	seen := map[string]bool{}
	for _, p := range f.rows {
		host := projectOwner(p)
		name := p.RemoteHost
		if name == "" {
			name = "this device"
		}
		if !seen[host] && f.olderPages[host] {
			turns = append(turns, projectTurn{host, "", "Newest projects on " + name})
		}
		seen[host] = true
		if p.NextCursor != "" {
			turns = append(turns, projectTurn{host, p.NextCursor, "Older projects on " + name})
		}
	}
	return turns
}
func (g *Game) turnProjectPage(turn projectTurn) {
	f := &g.projects
	source, ok := f.store.(projectPager)
	if !ok {
		return
	}
	g.projectWork(func(ctx context.Context) projectResult {
		rows, err := source.ProjectPage(ctx, turn.host, turn.cursor)
		if err != nil {
			return projectResult{message: "Could not load this project page"}
		}
		return projectResult{applyField: func() {
			if len(rows) == 0 {
				f.message = "The catalog changed. Reopen Projects to refresh."
				return
			}
			combined := make([]domain.Project, 0, len(f.rows)+len(rows))
			replaced := false
			for _, p := range f.rows {
				if projectOwner(p) != turn.host {
					combined = append(combined, p)
				} else if !replaced {
					combined = append(combined, rows...)
					replaced = true
				}
			}
			f.rows = combined
			if f.olderPages == nil {
				f.olderPages = map[string]bool{}
			}
			f.olderPages[turn.host] = turn.cursor != ""
			f.workspace = workspacePanel{}
			delete(g.selectedSections, hud.TabProjects)
			g.showList(hud.TabProjects)
		}}
	})
}

func (g *Game) projectTitle(p domain.Project) string {
	host := p.RemoteHost
	if host == "" {
		host = g.serviceProjection.LocalDevice.DisplayName("")
		if g.serviceProjection.LocalDevice.ID == "" {
			host = g.serviceProjection.ProjectHostname
			if host == "" {
				host = strings.Join(strings.Fields(g.runtimeInfo.DeviceName), "-")
			}
			if host == "" {
				host, _ = os.Hostname()
			}
			if !domain.ValidDeviceName(host) {
				host = "local"
			}
		}
	}
	return host + "://" + p.Name
}
