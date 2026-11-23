//go:build gui

// Projects live local services into Cluster Routing without HTTP or SQL.
// Selection remains local and projections rebuild only when displayed data changes.
package gui

import (
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/hud"
)

func (game *Game) Close() {
	game.stopLAN()
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
		for _, tab := range game.shell.Snapshot().Tabs {
			if tab.ID() != hud.TabCluster {
				continue
			}
			old := game.presentedTab(tab)
			selected := game.selectedSectionIndex(old)
			game.serviceProjection = projection
			if selected >= 0 {
				delete(game.selectedSections, hud.TabCluster)
				for i, s := range game.presentedTab(tab).Sections {
					if s.ActionID == old.Sections[selected].ActionID {
						game.selectedSections[hud.TabCluster] = i
						break
					}
				}
			}
			break
		}
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
		tab.Summary = "Known devices are saved on this installation. Routing, tasks and inference types show examples."
		sections := []hud.Section{game.liveServicesSection()}
		for _, s := range tab.Sections {
			if s.Title == "Cluster Devices" && s.IsSelectorHeading() {
				switch {
				case game.serviceProjection.DeviceError:
					s.Description = "Could not load saved devices."
				case len(game.serviceProjection.Devices) == 0:
					s.Description = "Use Add New Device to find a worker."
				default:
					s.Description = "Previously authenticated workers."
				}
			}
			sections = append(sections, s)
			if s.ActionID == "add-new-device" {
				for _, d := range game.serviceProjection.Devices {
					sections = append(sections, hud.Section{ActionID: "worker-" + d.Fingerprint, Title: d.Host, SelectorColor: s.SelectorColor, ContentKind: hud.ContentStandard, Rows: []hud.Row{{Label: "Last known endpoint", Detail: d.Address}, {Label: "Worker fingerprint", Detail: d.Fingerprint}, {Label: "Cluster", Detail: d.Cluster}}})
				}
			}
		}
		tab.Sections = sections
	}
	return tab
}
