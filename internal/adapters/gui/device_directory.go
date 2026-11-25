//go:build gui

// Presents the last synchronized directory and routes edits to the reporting worker.
// The fingerprint identifies a device across multiple worker snapshots.
package gui

import (
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"slices"
	"strings"
)

func (g *Game) personalSections() []hud.Section {
	seen := map[string]bool{}
	var rows []hud.Section
	var peers []domain.PoolDevice
	for _, w := range g.serviceProjection.Devices {
		peers = append(peers, w.Roster...)
	}
	if g.serviceProjection.LocalDevice.ID != "" {
		peers = append(peers, g.serviceProjection.LocalDevice)
	}
	peers = append(peers, g.serviceProjection.Peers...)
	for _, d := range peers {
		if d.Worker || seen[d.ID] {
			continue
		}
		seen[d.ID] = true
		if d.ID == g.serviceProjection.LocalDevice.ID && !domain.ValidDeviceName(d.Host) {
			d = g.serviceProjection.LocalDevice
		}
		icon, platform := "1f5a7", "Platform not reported"
		switch d.OS {
		case "android", "ios":
			icon, platform = "1f4f1", d.OS
		case "darwin", "linux", "windows", "freebsd", "openbsd", "netbsd":
			icon, platform = "1f4bb", d.OS
		}
		rows = append(rows, hud.Section{ActionID: "gui-" + d.ID, Title: d.DisplayName(""), SelectorColor: hud.SelectorPalette[2], SelectorIcon: icon, CurrentDevice: d.ID != "" && d.ID == g.serviceProjection.LocalDevice.ID, ContentKind: hud.ContentStandard, Rows: []hud.Row{{Label: "App device", Detail: platform}, {Label: "Device fingerprint", Detail: d.ID}}})
	}
	slices.SortStableFunc(rows, func(a, b hud.Section) int {
		return strings.Compare(strings.ToLower(a.Title), strings.ToLower(b.Title))
	})
	return rows
}

func (g *Game) deviceManagementLinks(action string) []domain.ClusterOption {
	var options []domain.ClusterOption
	if action == "cluster-status" {
		seen := map[string]bool{}
		for _, d := range g.serviceProjection.Devices {
			if d.Admin && !seen[d.Cluster] {
				seen[d.Cluster] = true
				options = append(options, domain.ClusterOption{Label: "Inference validators · " + d.DisplayName(), Route: domain.ClusterRoute{Page: "inference-roster", Target: d.Cluster}})
			}
		}
	}
	for _, w := range g.serviceProjection.Devices {
		if !w.Admin {
			continue
		}
		for _, d := range w.Roster {
			id := "gui-" + d.ID
			if d.Worker {
				id = "worker-" + d.ID
			}
			if id == action {
				if d.Worker {
					options = append(options, domain.ClusterOption{Label: "Select pools", Route: domain.ClusterRoute{Worker: w.Fingerprint, Page: "device-pools", Target: d.ID}})
				}
				page := "rename"
				if action == "gui-"+g.serviceProjection.LocalDevice.ID {
					page = "rename-local"
				}
				label := "Rename " + d.DisplayName(w.Host)
				if len(g.serviceProjection.Devices) > 1 {
					label += " via " + w.DisplayName()
				}
				options = append(options, domain.ClusterOption{Label: label, Route: domain.ClusterRoute{Worker: w.Fingerprint, Page: page, Target: d.ID}})
			}
		}
	}
	local := g.serviceProjection.LocalDevice
	if len(options) == 0 && local.ID != "" && action == "gui-"+local.ID {
		options = append(options, domain.ClusterOption{Label: "Rename " + local.DisplayName(""), Route: domain.ClusterRoute{Page: "rename-local", Target: local.ID}})
	}
	if local.ID != "" && action == "gui-"+local.ID {
		options = append(options, domain.ClusterOption{Label: "Performance", Route: domain.ClusterRoute{Page: "performance"}})
	}
	for _, w := range g.serviceProjection.Devices {
		if w.Admin && action == "worker-"+w.Fingerprint {
			opt := g.ollamaOption(w.Fingerprint)
			if opt.Label != "" {
				options = append(options, opt)
			}
		}
	}
	if action == "gui-"+local.ID && local.ID != "" && g.serviceProjection.LocalInventory.Observed > 0 {
		opt := g.ollamaOption("local")
		if opt.Label != "" {
			options = append(options, opt)
		}
	}
	return options
}

func (g *Game) routingMembers(section hud.Section) hud.Section {
	for _, pool := range domain.DefaultPools() {
		if section.ActionID != strings.ReplaceAll(strings.ToLower(pool.Name), " ", "-") {
			continue
		}
		return g.poolSection(section, pool.Name, pool.Priority)
	}
	return section
}

func (g *Game) poolSection(section hud.Section, name, priority string) hud.Section {
	section.Description = "Devices assigned to this pool at the last synchronization."
	section.Rows = nil
	if priority != "" {
		section.Rows = append(section.Rows, hud.Row{Label: "Priority", Detail: priority})
	}
	seen := map[string]bool{}
	for _, w := range g.serviceProjection.Devices {
		for _, d := range w.Roster {
			if d.Worker && !seen[d.ID] && slices.Contains(d.Pools, name) {
				section.Rows = append(section.Rows, hud.Row{Label: "Device", Detail: d.DisplayName(w.Host)})
				seen[d.ID] = true
			}
		}
	}
	if len(seen) == 0 {
		section.Rows = append(section.Rows, hud.Row{Label: "Devices", Detail: "No synchronized assignments."})
	}
	return section
}

func (g *Game) customPools() []hud.Section {
	seen := map[string]bool{}
	for _, p := range domain.DefaultPools() {
		seen[p.Name] = true
	}
	var names []string
	for _, w := range g.serviceProjection.Devices {
		for _, name := range w.PoolNames {
			if !seen[name] && !domain.RetiredPool(name) {
				seen[name] = true
				names = append(names, name)
			}
		}
	}
	slices.Sort(names)
	var sections []hud.Section
	for _, name := range names {
		section := hud.Section{Title: name, ActionID: "routing-pool-" + name, SelectorColor: hud.SelectorPalette[1], ContentKind: hud.ContentStandard}
		sections = append(sections, g.poolSection(section, name, ""))
	}
	return sections
}

func (g *Game) discoveryLabel(host string) string {
	for _, w := range g.serviceProjection.Devices {
		if strings.EqualFold(w.Host, host) {
			return w.DisplayName()
		}
	}
	if domain.ValidDeviceName(host) {
		return host
	}
	return "Unnamed device"
}
