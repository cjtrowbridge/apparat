//go:build gui

// Projects live local services into Cluster Routing without HTTP or SQL.
// Selection remains local and projections rebuild only when displayed data changes.
package gui

import (
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/hud"
)

func (game *Game) Close() {
	game.closeDesktopUpdates()
	if game.projects.close != nil {
		game.projects.close()
		game.projects.workers.Wait()
	}
	if game.voice.engine != nil {
		game.voice.engine.Close()
	}
	game.flushPlacement()
	if game.cancelQueries != nil {
		game.cancelQueries()
		game.cancelQueries = nil
	}
}
func (game *Game) receiveServices() {
	select {
	case projection, ok := <-game.serviceUpdates:
		if !ok {
			game.serviceUpdates = nil
			return
		}
		game.serviceProjection = projection
		if game.shell.Snapshot().ActiveTab().ID() == hud.TabCluster {
			game.layoutDirty = true
		}
	default:
	}
}
func (game *Game) liveServicesSection() hud.Section {
	section := hud.Section{ActionID: "local-services", Title: "Local services", Description: game.serviceProjection.State}
	for _, service := range game.serviceProjection.Services {
		section.Rows = append(section.Rows, hud.Row{Label: service.Name, Detail: fmt.Sprintf("%s | %s | %s | revision %d", service.ID, service.Driver, service.Health, service.Revision)})
	}
	if len(section.Rows) == 0 {
		section.Rows = []hud.Row{{Label: "Services", Detail: game.serviceProjection.State}}
	}
	return section
}

func (game *Game) presentedTab(tab hud.Tab) hud.Tab {
	if tab.ID() == hud.TabProjects && game.projects.store != nil {
		return game.projectTab(tab)
	}
	if tab.ID() == hud.TabCluster && game.runtimeInfo.Queries != nil {
		tab.Summary = "Local services are live mock drivers. Devices, routing pools and tasks remain demonstrations."
		tab.Sections = append([]hud.Section{game.liveServicesSection()}, tab.Sections...)
	}
	return tab
}
