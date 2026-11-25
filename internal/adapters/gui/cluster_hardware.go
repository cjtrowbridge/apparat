//go:build gui

// Totals installed capacity across the current local and authenticated cluster directory.
// Metric captions give reporting coverage when only some hosts supplied a measurement.
package gui

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
)

func hardwareMembers(p ServiceProjection) map[string]domain.HardwareCapacity {
	devices := map[string]domain.HardwareCapacity{}
	add := func(id string, h domain.HardwareCapacity) {
		if id == "" {
			return
		}
		_, exists := devices[id]
		if !exists || h != (domain.HardwareCapacity{}) {
			devices[id] = h
		}
	}
	// Consult cached capacity only for current members of the reporting directory.
	for _, d := range p.Devices {
		for id, h := range d.RosterHardware {
			add(id, h)
		}
		add(d.Fingerprint, d.Hardware[d.Fingerprint])
		for _, peer := range d.Roster {
			add(peer.ID, d.Hardware[peer.ID])
		}
	}
	for _, d := range p.Devices {
		add(d.Fingerprint, domain.HardwareCapacity{})
		if d.Inventory != nil && d.Inventory.Observed > 0 {
			add(d.Fingerprint, d.Inventory.Capacity())
		}
	}
	for _, peer := range p.Peers {
		add(peer.ID, domain.HardwareCapacity{})
	}
	local := p.LocalDevice.ID
	if local == "" && p.LocalInventory.Observed > 0 {
		local = "local"
	}
	add(local, p.LocalInventory.Capacity())
	return devices
}

func hardwareCard(p ServiceProjection, _ time.Time) hud.Section {
	devices := hardwareMembers(p)
	hosts := map[string]domain.HardwareCapacity{}
	missing := 0
	for id, h := range devices {
		if h == (domain.HardwareCapacity{}) {
			missing++
			continue
		}
		key := id
		if h.Boot != "" {
			key = "boot/" + h.Boot
		}
		prior := hosts[key]
		// One physical host may report different counters through its GUI and worker.
		h.Cores = max(prior.Cores, h.Cores)
		h.Threads = max(prior.Threads, h.Threads)
		h.RAM = max(prior.RAM, h.RAM)
		h.VRAM = max(prior.VRAM, h.VRAM)
		h.Disk = max(prior.Disk, h.Disk)
		h.INT8 = max(prior.INT8, h.INT8)
		rated := max(prior.GPUs-prior.UnratedGPUs, h.GPUs-h.UnratedGPUs)
		h.GPUs = max(prior.GPUs, h.GPUs, prior.UnratedGPUs, h.UnratedGPUs)
		h.UnratedGPUs = h.GPUs - rated
		hosts[key] = h
	}
	card := hud.Section{ContentKind: "cluster-hardware", Title: "Cluster hardware", Description: fmt.Sprintf("%d reporting hosts", len(hosts))}
	if len(hosts) == 1 {
		card.Description = "1 reporting host"
	}
	if missing > 0 {
		card.Description += fmt.Sprintf(" · %d devices awaiting hardware reports", missing)
	}
	if len(hosts) == 0 {
		return card
	}
	var totals [6]float64
	var reports [6]int
	unknownGPU, gpuCount := 0, 0
	for _, h := range hosts {
		values := [6]float64{float64(h.Cores), float64(h.Threads), float64(h.RAM), float64(h.VRAM), float64(h.Disk), h.INT8}
		unknownGPU += h.UnratedGPUs
		gpuCount += max(h.GPUs, h.UnratedGPUs)
		for i, value := range values {
			if value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0) {
				totals[i] += value
				reports[i]++
			}
		}
	}
	if gpuCount > 0 {
		card.Description += fmt.Sprintf(" · %d GPUs", gpuCount)
	}
	if unknownGPU > 0 {
		card.Description += fmt.Sprintf(" · %d GPU(s) awaiting compute measurements", unknownGPU)
	}
	for i, label := range []string{"CPU cores", "CPU threads", "RAM", "GPU VRAM", "Disk capacity", "Est. INT8 compute"} {
		value := "—"
		if reports[i] > 0 {
			switch i {
			case 0, 1:
				value = fmt.Sprintf("%.0f", totals[i])
			case 5:
				value = hardwareUnits(totals[i], 1000, []string{"OP/s", "kOP/s", "MOP/s", "GOP/s", "TOP/s", "POP/s"})
			default:
				value = hardwareUnits(totals[i], 1024, []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB"})
			}
		}
		if reports[i] > 0 && reports[i] < len(hosts) {
			label += fmt.Sprintf(" (%d/%d)", reports[i], len(hosts))
		}
		card.Rows = append(card.Rows, hud.Row{Label: label, Detail: value})
	}
	return card
}
func hardwareUnits(value, base float64, units []string) string {
	i := 0
	for value >= base && i < len(units)-1 {
		value /= base
		i++
	}
	rounded := math.Round(value*10) / 10
	return strconv.FormatFloat(rounded, 'f', -1, 64) + " " + units[i]
}
func (g *Game) hardwareGrid(rows []hud.Row) *widget.Container {
	width := g.detailPanePreferredWidth() - 64
	widest := 1
	for _, r := range rows {
		number, unit, _ := strings.Cut(r.Detail, " ")
		unitWidth, _ := text.Measure(unit, *g.formLabelFace, 0)
		w, _ := text.Measure(number, *g.pageTitleFace, 0)
		h, _ := text.Measure(r.Label, *g.formLabelFace, 0)
		widest = max(widest, int(math.Ceil(max(w+unitWidth+4, h))))
	}
	columns := max(1, min(4, (width+12)/(widest+12)))

	grid := widget.NewContainer(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})), widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(columns), widget.GridLayoutOpts.DefaultStretch(true, false), widget.GridLayoutOpts.Spacing(12, 8))))
	for _, r := range rows {
		cell := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))))
		number, unit, _ := strings.Cut(r.Detail, " ")
		value := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionHorizontal), widget.RowLayoutOpts.Spacing(4))))
		value.AddChild(widget.NewText(widget.TextOpts.Text(number, g.pageTitleFace, g.theme.ButtonTheme.TextColor.Idle)))
		if unit != "" {
			value.AddChild(widget.NewText(widget.TextOpts.Text(unit, g.formLabelFace, g.theme.ButtonTheme.TextColor.Idle), widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Position: widget.RowLayoutPositionCenter}))))
		}
		cell.AddChild(value, widget.NewText(widget.TextOpts.Text(r.Label, g.formLabelFace, mutedTextColor)))

		grid.AddChild(cell)
	}
	return grid
}
