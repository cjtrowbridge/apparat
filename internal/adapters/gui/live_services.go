//go:build gui

// Projects live local services into Cluster Routing without HTTP or SQL.
// Selection remains local and projections rebuild only when displayed data changes.
package gui

import (
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"maps"
	"os"
	"slices"
	"strings"
	"time"
)

func (game *Game) Close() {
	game.preview.stop()
	for _, copy := range game.projects.externalCopies {
		os.RemoveAll(copy.dir)
	}
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
		hostChanged := game.serviceProjection.ProjectHostname != projection.ProjectHostname || game.serviceProjection.LocalDevice.Host != projection.LocalDevice.Host
		nameChanged := game.serviceProjection.ScreenName != projection.ScreenName
		for _, tab := range game.shell.Snapshot().Tabs {
			if tab.ID() != hud.TabCluster {
				continue
			}
			old := game.presentedTab(tab)
			selected := game.selectedSectionIndex(old)
			game.serviceProjection = projection
			if game.lan.selected != "" && !slices.Contains(game.discoverableNames(), game.lan.selected) {
				game.leavePeer()
			}
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
		active := game.shell.Snapshot().ActiveTab().ID()
		if active == hud.TabCluster || active == hud.TabSettings && nameChanged || active == hud.TabProjects && hostChanged {
			game.layoutDirty = true
		}
	default:
	}
}
func (game *Game) liveServicesSection() hud.Section {
	section := hud.Section{ActionID: "cluster-status", Title: "Cluster Status"}
	agreement := hud.Section{Title: "Claim agreement"}
	queue := taskStatusCard(game.serviceProjection.Devices)
	providers := hud.Section{Title: "Text inference providers"}
	for _, d := range game.serviceProjection.Devices {
		name := d.DisplayName()
		health, model := "Worker status unavailable", "Awaiting inventory"
		if d.Inventory != nil {
			switch {
			case !d.Inventory.Fresh(time.Now()):
				model = "Inventory stale"
			case d.Inventory.HasText(time.Now()):
				model = "Text generation available"
			default:
				model = "Text generation unavailable"
			}
		}
		if d.Inventory != nil && d.Inventory.Retrying {
			health = workerRetryMessage(name, d.Inventory, time.Now())
			model = health
			if d.Inventory.ReportedText() {
				model = "Text generation last reported · " + health
			}
		} else if d.Inference != nil && d.Inference.OfflineSince > 0 {
			health = "Status unavailable since " + time.UnixMilli(d.Inference.OfflineSince).Local().Format("15:04:05")
		} else if d.Inference != nil {
			v := d.Inference
			health = fmt.Sprintf("%d validators · confirmation %s · propagation %s · recheck %s · %d incomplete rounds", v.Validators, timingLabel(v.Timing.Confirmation), timingLabel(v.Timing.Propagation), v.Timing.Recheck, v.Timing.Incomplete)
			if v.Validators == 0 {
				health = "Synchronizing authenticated worker membership"
			}
			if v.Updating != "" {
				health += " · worker update in progress"
			}
			if v.Busy {
				model += " · model slot busy"
			} else if v.BlockedSince > 0 && v.Queued > 0 {
				model += " · idle with eligible work since " + time.UnixMilli(v.BlockedSince).Local().Format("15:04:05")
			} else {
				model += " · model slot idle"
			}
			for _, d := range game.serviceProjection.Devices {
				if slices.Contains(v.Awaiting, d.Fingerprint) {
					health += " · awaiting " + d.DisplayName()
					continue
				}
				o, ok := v.Peers[d.Fingerprint]
				if ok && !o.Received {
					health += " · missing reply from " + d.DisplayName()
				}
			}
		}
		agreement.Rows = append(agreement.Rows, hud.Row{Label: name, Detail: health})
		providers.Rows = append(providers.Rows, hud.Row{Label: name, Detail: model})
	}
	if len(game.serviceProjection.Devices) == 0 {
		section.Description = "No saved inference workers."
	}
	section.DetailSections = []hud.Section{hardwareCard(game.serviceProjection, time.Now()), queue, deviceUsageCard(game.serviceProjection), agreement, providers}
	return section
}

func timingLabel(s domain.TimingSummary) string {
	if s.Count == 0 {
		return "awaiting samples"
	}
	return fmt.Sprintf("p99 %s (%d samples)", s.P99.Round(time.Microsecond), s.Count)
}

func (game *Game) presentedTab(tab hud.Tab) hud.Tab {
	if tab.ID() == hud.TabProjects && game.projects.store != nil {
		return game.projectTab(tab)
	}
	if tab.ID() == hud.TabCluster && game.runtimeInfo.Queries != nil {
		tab.Summary = "Browse saved devices and pool assignments. Inference types show current worker advertisements."
		sections := []hud.Section{game.liveServicesSection()}
		for _, s := range tab.Sections {
			if s.Title == "Cluster Devices" && s.IsSelectorHeading() {
				switch {
				case game.serviceProjection.DeviceError:
					s.Description = "Could not load saved devices."
				case len(game.serviceProjection.Devices) == 0 && len(game.serviceProjection.Peers) == 0 && game.serviceProjection.LocalDevice.ID == "":
					s.Description = "Use Add New Device to find a worker."
				default:
					s.Description = "Personal app devices appear above headless workers."
				}
			}
			if s.Title == "Tasks" && s.IsSelectorHeading() {
				sections = append(sections, game.customPools()...)
			}
			s = game.routingMembers(s)
			sections = append(sections, s)
			if s.ActionID == "add-new-device" {
				sections = append(sections, game.personalSections()...)
				for _, d := range game.serviceProjection.Devices {
					rows := []hud.Row{{Label: "Last known endpoint", Detail: d.Address}, {Label: "Worker fingerprint", Detail: d.Fingerprint}, {Label: "Cluster", Detail: d.Cluster}}
					if d.ReplicaRevision > 0 {
						rows = append(rows, hud.Row{Label: "Last applied worker view", Detail: fmt.Sprintf("Revision %d · %d available pools · checks every %d seconds", d.ReplicaRevision, d.PoolCount, d.PollSeconds)})
						if d.Admin {
							rows = append(rows, hud.Row{Label: "Worker policy at last sync", Detail: fmt.Sprintf("Revision %d", d.PolicyRevision)})
						}
					} else {
						rows = append(rows, hud.Row{Label: "Synchronization", Detail: "Awaiting first authorized worker view."})
					}
					sections = append(sections, hud.Section{ActionID: "worker-" + d.Fingerprint, Title: d.DisplayName(), SelectorColor: s.SelectorColor, ContentKind: hud.ContentStandard, Rows: rows})
				}
			}
		}
		tab.Sections = game.inventorySections(sections)
	}
	return tab
}

func (g *Game) workerAction(action string, callback func(string)) func() {
	for _, d := range g.serviceProjection.Devices {
		if callback != nil && action == "worker-"+d.Fingerprint {
			return func() {
				go callback(d.Fingerprint)
			}
		}
	}
	return nil
}

func taskStatusCard(devices []domain.ClusterDevice) hud.Section {
	card := hud.Section{Title: "Inference tasks", Description: "Latest 32 tasks visible to you across the cluster."}
	tasks := map[string]domain.TaskObservation{}
	turns := map[string]bool{}
	rank := map[string]int{"received": 1, "queued": 2, "owned": 3, "dispatching": 4, "finished": 5}
	reported := false
	for _, d := range devices {
		if d.Inference == nil {
			continue
		}
		if d.Inference.OfflineSince == 0 {
			reported = true
		}
		for _, t := range d.Inference.Recent {
			if t.State != "received" {
				turns[t.Turn] = true
			}
			if rank[t.State] > rank[tasks[t.ID].State] {
				tasks[t.ID] = t
			}
		}
	}
	ids := slices.Sorted(maps.Keys(tasks))
	slices.Reverse(ids)
	for _, id := range ids {
		t := tasks[id]
		if t.State == "received" && turns[t.Turn] {
			continue
		}
		card.Rows = append(card.Rows, hud.Row{Label: t.Pool + " · " + t.State, Detail: t.ID})
		if len(card.Rows) == 32 {
			break
		}
	}
	if len(card.Rows) == 0 {
		card.Description = "Waiting for a worker task report."
		if reported {
			card.Description = "No recent tasks reported for your account."
		}
	}
	return card
}

func deviceUsageCard(p ServiceProjection) hud.Section {
	card := hud.Section{Title: "Devices", Description: "Installed hardware and lifetime inference totals."}
	members := hardwareMembers(p)
	names := map[string]string{}
	details := map[string]string{}
	indirect := map[string]domain.DeviceUsage{}
	for _, d := range p.Devices {
		for id, usage := range d.RosterUsage {
			prior, ok := indirect[id]
			if !ok || usage.Prompts > prior.Prompts {
				indirect[id] = usage
			}
		}
	}
	for id, usage := range indirect {
		details[id] = fmt.Sprintf("%d tokens generated · %d prompts served", usage.Tokens, usage.Prompts)
	}
	for _, peer := range p.Peers {
		names[peer.ID] = peer.DisplayName("")
	}
	names[p.LocalDevice.ID] = p.LocalDevice.DisplayName("")
	for _, d := range p.Devices {
		for _, peer := range d.Roster {
			names[peer.ID] = peer.DisplayName(d.Host)
		}
	}
	for _, d := range p.Devices {
		names[d.Fingerprint] = d.DisplayName()
		usage := d.Usage
		if d.Inventory != nil && d.Inventory.Usage != nil && (usage == nil || usage.Prompts < d.Inventory.Usage.Prompts) {
			usage = d.Inventory.Usage
		}
		if usage != nil {
			details[d.Fingerprint] = fmt.Sprintf("%d tokens generated · %d prompts served", usage.Tokens, usage.Prompts)
		}
		if d.Inventory != nil && d.Inventory.Retrying {
			details[d.Fingerprint] = strings.TrimSpace(details[d.Fingerprint] + " " + workerRetryMessage(d.DisplayName(), d.Inventory, time.Now()))
		}
	}
	ids := slices.Sorted(maps.Keys(members))
	for _, id := range ids {
		h := members[id]
		name := names[id]
		if name == "" {
			name = "Device " + id[:min(8, len(id))]
		}
		desc := "Awaiting hardware report"
		if h != (domain.HardwareCapacity{}) {
			desc = fmt.Sprintf("%d CPU threads", h.Threads)
			if h.GPUs > 0 || h.UnratedGPUs > 0 {
				desc += fmt.Sprintf(" · %d GPUs", max(h.GPUs, h.UnratedGPUs))
			}
			if h.UnratedGPUs > 0 {
				desc += fmt.Sprintf(" (%d awaiting compute measurements)", h.UnratedGPUs)
			}
		}
		detail := details[id]
		if detail == "" {
			detail = "Inference totals unreported"
		}
		card.Rows = append(card.Rows, hud.Row{Label: name, Detail: desc + " · " + detail})
	}
	slices.SortStableFunc(card.Rows, func(a, b hud.Row) int {
		return strings.Compare(a.Label, b.Label)
	})
	return card
}
