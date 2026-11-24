//go:build gui

// Presents measured device capacity, Ollama inventory and future capability placeholders.
// The selector badge and per-type device list derive from the same fresh observation.
package gui

import (
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
	"strings"
	"time"
)

var inferenceKinds = []struct{ id, title, icon string }{
	{"text-generation", "Text Generation", "1f4ac"},
	{"image-generation", "Image Generation", "1f5bc"},
	{"text-to-speech", "Text-to-Speech", "1f50a"},
	{"speech-to-text", "Speech-to-Text", "1f3a4"},
	{"video-generation", "Video Generation", "1f3ac"},
}

func (g *Game) ollamaOption(worker string) domain.ClusterOption {
	var v *domain.WorkerInventory
	if worker == "local" {
		v = &g.serviceProjection.LocalInventory
	} else {
		for _, d := range g.serviceProjection.Devices {
			if d.Fingerprint == worker {
				v = d.Inventory
				break
			}
		}
	}
	label := "Configure Ollama port"
	if v != nil && (v.Configured || v.Version != "") {
		label = "Change Ollama port"
	}
	return domain.ClusterOption{Label: label, Route: domain.ClusterRoute{Page: "ollama", Worker: worker}}
}
func inventoryDetail(v *domain.WorkerInventory, configure string) []hud.Section {
	now := time.Now()
	capacity := hud.Section{Title: "System resources", Description: "No authenticated resource report available."}
	if v != nil && v.Observed > 0 {
		r := v.Resources
		capacity.Description = "Observed " + time.Unix(v.Observed, 0).Local().Format("2006-01-02 15:04:05 MST")
		if !v.Fresh(now) {
			capacity.Description += " · stale"
		}
		cpu := fmt.Sprintf("%d logical CPUs · %s/%s", r.CPUs, r.OS, r.Arch)
		if r.CPUPercent != nil {
			cpu += fmt.Sprintf(" · %.0f%% used", *r.CPUPercent)
		}
		capacity.Rows = append(capacity.Rows, hud.Row{Label: "Processor", Detail: cpu})
		add := func(label string, c *domain.Capacity) {
			detail := "Unavailable"
			if c != nil {
				const gib = 1 << 30
				detail = fmt.Sprintf("%.2f GiB used · %.2f GiB available · %.2f GiB total", float64(c.Used)/gib, float64(c.Available)/gib, float64(c.Total)/gib)
			}
			capacity.Rows = append(capacity.Rows, hud.Row{Label: label, Detail: detail})
		}
		add("RAM", r.RAM)
		add("Swap", r.Swap)
		if len(r.GPUs) == 0 {
			add("VRAM", nil)
		}
		for _, g := range r.GPUs {
			add(g.Name+" VRAM", &g)
		}
		if len(r.Disks) == 0 {
			add("Disk", nil)
		}
		for _, d := range r.Disks {
			add(d.Name, &d)
		}
	}
	details := []hud.Section{capacity}
	for _, kind := range inferenceKinds {
		s := hud.Section{Title: kind.title, Description: "Unavailable · discovery is not implemented."}
		if kind.id == "text-generation" {
			s.Description = "Awaiting the worker’s Ollama inventory."
			if v != nil && v.Configured {
				s.Description = "Ollama configured on this device · awaiting a current health report."
			}
			if configure != "" {
				s.ActionID = "configure-ollama/" + configure
			}
			if v != nil && v.Fresh(now) {
				if v.Problem != "" {
					s.Description = v.Problem
					if v.Configured {
						s.Description = "Ollama configured · last health check failed. " + v.Problem
					}
				} else if v.Version != "" {
					s.Description = "Ollama " + v.Version + " · configured"
					if !v.HasText(now) {
						s.Description += " · no text-generation models reported"
					}
					for _, model := range v.Models {
						status := "Installed"
						if model.Loaded {
							status = "Loaded"
						}
						if model.Text {
							status += " · text generation"
						}
						s.Rows = append(s.Rows, hud.Row{Label: model.Name, Detail: status + fmt.Sprintf(" · %.2f GiB", float64(model.Size)/(1<<30))})
					}
				}
			}
		}
		details = append(details, s)
	}
	return details
}
func (g *Game) inventorySections(sections []hud.Section) []hud.Section {
	type source struct {
		name, configure string
		v               *domain.WorkerInventory
	}
	devices := map[string]source{}
	local := g.serviceProjection.LocalDevice
	localConfig := ""
	if g.serviceProjection.LocalInventory.Observed > 0 {
		localConfig = "local"
	}
	devices["gui-"+local.ID] = source{local.DisplayName(""), localConfig, &g.serviceProjection.LocalInventory}
	for _, d := range g.serviceProjection.Devices {
		config := ""
		if d.Admin {
			config = d.Fingerprint
		}
		devices["worker-"+d.Fingerprint] = source{d.DisplayName(), config, d.Inventory}
	}
	var textRows []hud.Row
	for i := range sections {
		s := &sections[i]
		d, ok := devices[s.ActionID]
		if ok || len(s.ActionID) > 4 && s.ActionID[:4] == "gui-" {
			s.DetailSections = inventoryDetail(d.v, d.configure)
			if d.v != nil && d.v.HasText(time.Now()) {
				s.CapabilityIcon = inferenceKinds[0].icon
				textRows = append(textRows, hud.Row{Label: d.name, Detail: "Ollama " + d.v.Version})
			}
		}
	}
	for i := range sections {
		s := &sections[i]
		for _, k := range inferenceKinds {
			if k.id != s.ActionID {
				continue
			}
			s.CenterIcon = k.icon
			s.Rows = nil
			s.Description = "Unavailable · discovery is not implemented."
			if k.id == "text-generation" {
				s.Description = "No device currently advertises text generation."
				if len(textRows) > 0 {
					s.Description = "Devices advertising text generation."
					s.Rows = textRows
				}
			}
		}
	}
	return sections
}

func (g *Game) addInventoryControl(c *widget.Container, action string) {
	worker, ok := strings.CutPrefix(action, "configure-ollama/")
	if ok {
		opt := g.ollamaOption(worker)
		g.projectButton(c, opt.Label, func() {
			g.clusterOpen(opt.Route)
		})
	}
}

func (g *Game) applyInventoryFocus() {
	f := g.clusterTop()
	if f == nil || !f.focusInput || f.input == nil || g.nav.owner != hud.TabCluster || g.actions.mode != "" {
		return
	}
	targets := g.bodyTargets()
	for i, target := range targets {
		if target.w == f.input {
			g.focusBody(targets, i)
			f.focusInput = false
			return
		}
	}
}

// The scroll pane owns each card so its background and gutters span the same width.
func (g *Game) addInventoryCards(pane *widget.Container, sections []hud.Section) {
	for _, s := range sections {
		card := g.detailCard()
		card.AddChild(widget.NewText(widget.TextOpts.Text(s.Title, g.pageTitleFace, g.theme.ButtonTheme.TextColor.Idle)))
		card.AddChild(g.detailText(s.Description))
		g.addInventoryControl(card, s.ActionID)
		for _, row := range s.Rows {
			label := wrapSelectorLabel(row.Label, *g.formLabelFace, int(g.detailTextMaxWidth()))
			card.AddChild(widget.NewText(widget.TextOpts.Text(label, g.formLabelFace, g.theme.ButtonTheme.TextColor.Idle)))
			card.AddChild(widget.NewText(widget.TextOpts.Text(row.Detail, g.theme.ButtonTheme.TextFace, mutedTextColor), widget.TextOpts.MaxWidth(g.detailTextMaxWidth())))
		}
		pane.AddChild(card)
	}
}
