//go:build gui

// Builds a screenshot-friendly scorecard from fresh local and authenticated worker inventory.
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

func hardwareCard(p ServiceProjection, now time.Time) hud.Section {
	devices := map[string]domain.WorkerInventory{}
	add := func(id string, v *domain.WorkerInventory) {
		if id == "" {
			return
		}
		current := devices[id]
		if v != nil && v.Fresh(now) && v.Observed > current.Observed {
			current = *v
		}
		devices[id] = current
	}
	local := p.LocalDevice.ID
	if local == "" && p.LocalInventory.Observed != 0 {
		local = "local"
	}
	add(local, &p.LocalInventory)
	for _, peer := range p.Peers {
		add(peer.ID, nil)
	}
	for _, d := range p.Devices {
		add(d.Fingerprint, d.Inventory)
	}
	hosts := map[string]domain.WorkerInventory{}
	missing := 0
	for id, v := range devices {
		if v.Observed == 0 {
			missing++
			continue
		}
		key := id
		if v.Resources.HostBootDigest != "" {
			key = "boot/" + v.Resources.HostBootDigest
		}
		prior := hosts[key]
		if prior.Observed < v.Observed || prior.Observed == v.Observed && prior.Resources.CPUFP32PerSecond < v.Resources.CPUFP32PerSecond {
			hosts[key] = v
		}
	}
	card := hud.Section{ContentKind: "cluster-hardware", Title: "Cluster hardware", Description: fmt.Sprintf("%d reporting hosts", len(hosts))}
	if len(hosts) == 1 {
		card.Description = "1 reporting host"
	}
	if missing > 0 {
		card.Description += fmt.Sprintf(" · %d devices without fresh inventory", missing)
	}
	if len(hosts) == 0 {
		return card
	}
	var totals [6]float64
	var reports [6]int
	for _, v := range hosts {
		r := v.Resources
		values := [6]float64{float64(r.PhysicalCores), float64(r.CPUs)}
		if r.RAM != nil {
			values[2] = float64(r.RAM.Total)
		}
		for _, gpu := range r.GPUs {
			values[3] += float64(gpu.Total)
		}
		for _, disk := range r.Disks {
			values[4] += float64(disk.Total)
		}
		values[5] = r.CPUFP32PerSecond
		for i, value := range values {
			if value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0) {
				totals[i] += value
				reports[i]++
			}
		}
	}
	for i, label := range []string{"CPU cores", "CPU threads", "RAM", "GPU VRAM", "Disk capacity", "CPU FP32"} {
		value := "—"
		if reports[i] > 0 {
			switch i {
			case 0, 1:
				value = fmt.Sprintf("%.0f", totals[i])
			case 5:
				value = hardwareUnits(totals[i], 1000, []string{"FLOP/s", "kFLOP/s", "MFLOP/s", "GFLOP/s", "TFLOP/s", "PFLOP/s"})
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
	columns := max(1, min(3, (width+12)/(widest+12)))

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
