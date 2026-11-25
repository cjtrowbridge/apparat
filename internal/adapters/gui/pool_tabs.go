//go:build gui

// Presents pool tabs and preserves their filters when opening a device or task.
// Worker rows use the same detail renderer as the Cluster Devices list.
package gui

import (
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
	"slices"
	"strings"
	"time"
)

var poolViews = []struct{ name, icon string }{{"Tasks", "2611"}, {"Devices", "1f5a7"}, {"Authorized users", "1f465"}}
var taskFilters = []string{"Completed", "Queued", "Running", "Cancelled", "All"}

func sectionPool(s hud.Section) string {
	for _, p := range domain.DefaultPools() {
		if s.ActionID == strings.ReplaceAll(strings.ToLower(p.Name), " ", "-") {
			return p.Name
		}
	}
	if strings.HasPrefix(s.ActionID, "routing-pool-") {
		return strings.TrimPrefix(s.ActionID, "routing-pool-")
	}
	return ""
}
func (g *Game) poolDeviceDetail(c *widget.Container, id string) {
	for _, tab := range g.shell.Snapshot().Tabs {
		if tab.ID() != hud.TabCluster {
			continue
		}
		for _, s := range g.presentedTab(tab).Sections {
			if s.ActionID == id {
				g.sectionDetail(c, hud.TabCluster, s)
				return
			}
		}
	}
	c.AddChild(g.detailText("Device no longer available."))
}
func controlRow(c *widget.Container) *widget.Container {
	row := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionHorizontal), widget.RowLayoutOpts.Spacing(4))))
	c.AddChild(row)
	return row
}
func (g *Game) selectPoolView(f *clusterFrame, name string) {
	f.poolTab = name
	f.choosingFilter = false
	f.scroll = 0
	g.layoutDirty = true
}
func (g *Game) poolTabs(c *widget.Container, f *clusterFrame) {
	p := f.page.Overview
	row := controlRow(c)
	active := f.poolTab
	if active == "" {
		active = "Tasks"
	}
	for _, v := range poolViews {
		b := g.hudIcon(v.icon, func() {
			g.selectPoolView(f, v.name)
		})
		b.ToggleMode = true
		if v.name == active {
			b.SetState(widget.WidgetChecked)
		}
		row.AddChild(b)
	}
	c.AddChild(g.detailText(active))
	switch active {
	case "Devices":
		for _, opt := range p.Devices {
			g.projectButton(c, opt.Label, func() {
				g.clusterOpen(opt.Route)
			})
		}
		if len(p.Devices) == 0 {
			c.AddChild(g.detailText("No workers assigned."))
		}
	case "Authorized users":
		add := g.hudIcon("2795", func() {
			g.openPoolAdd(f)
		})
		add.GetWidget().Disabled = len(p.Add) == 0
		c.AddChild(add)
		for _, name := range p.Users {
			c.AddChild(g.detailText(name))
		}
		if len(p.Users) == 0 {
			c.AddChild(g.detailText("No authorized users visible."))
		}
	default:
		g.poolTaskRows(c, f)
	}
	for _, opt := range p.Policy {
		if active == "Authorized users" {
			g.projectButton(c, opt.Label, func() {
				g.clusterOpen(opt.Route)
			})
		}
	}
	g.projectButton(c, "Refresh", g.clusterLoad)
}
func (g *Game) openPoolAdd(f *clusterFrame) {
	opts := f.page.Overview.Add
	if len(opts) == 1 {
		g.clusterOpen(opts[0].Route)
		return
	}
	g.clusterPush(clusterFrame{page: domain.ClusterPage{Title: "Choose a worker", Detail: f.route.Pool, Options: opts}})
}
func poolTaskMatches(state, filter string) bool {
	if filter == "" || filter == "Completed" {
		return state == "succeeded" || state == "finished"
	}
	if filter == "Running" && (state == "owned" || state == "dispatching") {
		return true
	}
	return filter == "All" || strings.EqualFold(filter, state)
}
func (g *Game) poolTaskRows(c *widget.Container, f *clusterFrame) {
	filter := f.taskFilter
	if filter == "" {
		filter = "Completed"
	}
	row := controlRow(c)
	g.projectButton(row, filter, func() {
		f.choosingFilter = !f.choosingFilter
		g.layoutDirty = true
	})
	order := "Newest first"
	if f.oldestFirst {
		order = "Oldest first"
	}
	g.projectButton(row, order, func() {
		f.oldestFirst = !f.oldestFirst
		g.layoutDirty = true
	})
	if f.choosingFilter {
		for _, choice := range taskFilters {
			g.projectButton(c, choice, func() {
				f.taskFilter = choice
				f.choosingFilter = false
				g.layoutDirty = true
			})
		}
	}
	tasks := f.page.Overview.Tasks
	if f.oldestFirst {
		tasks = slices.Clone(tasks)
		slices.Reverse(tasks)
	}
	count := 0
	for _, t := range tasks {
		if !poolTaskMatches(t.State, filter) {
			continue
		}
		count++
		label := time.UnixMilli(t.Changed).Local().Format("2006-01-02 15:04") + " · " + t.State + " · " + t.Host + " · " + t.ID
		g.projectButton(c, wrapSelectorLabel(label, *g.theme.ButtonTheme.TextFace, max(1, g.detailPanePreferredWidth()-52)), func() {
			g.clusterOpen(domain.ClusterRoute{Page: poolTaskPage(t), Pool: f.route.Pool, Worker: t.Worker, Target: t.ID})
		})
	}
	if count == 0 {
		c.AddChild(g.detailText("No matching tasks."))
	}
}

func (g *Game) poolActions(add registerAction, index int, f *clusterFrame) {
	p := f.page.Overview
	if p == nil {
		return
	}
	path := "Cluster / " + f.route.Pool
	owner := f.route.Worker + "/" + f.route.Pool
	nav := func(kind, id, name string, open func()) {
		add(kind, owner, id, name, path+" / "+name, func() {
			g.selectTab(index, nil)
			open()
		})
	}
	for _, v := range poolViews {
		nav("pool-tab", v.icon, v.name, func() {
			g.selectPoolView(f, v.name)
		})
	}
	switch f.poolTab {
	case "Devices":
		for _, o := range p.Devices {
			nav("pool-worker", o.Route.Target, o.Label, func() {
				g.clusterOpen(o.Route)
			})
		}
	case "Authorized users":
		if len(p.Add) > 0 {
			nav("pool-add-user", "add", "Add authorized user", func() {
				g.openPoolAdd(f)
			})
		}
		for _, o := range p.Policy {
			nav("pool-policy", o.Route.Worker, o.Label, func() {
				g.clusterOpen(o.Route)
			})
		}
	default:
		for _, filter := range taskFilters {
			nav("pool-filter", filter, filter+" tasks", func() {
				f.taskFilter = filter
				f.choosingFilter = false
				g.layoutDirty = true
			})
		}
		sort := "Oldest first"
		if f.oldestFirst {
			sort = "Newest first"
		}
		nav("pool-sort", sort, sort, func() {
			f.oldestFirst = !f.oldestFirst
			g.layoutDirty = true
		})
		for _, t := range p.Tasks {
			if poolTaskMatches(t.State, f.taskFilter) {
				nav("pool-task", t.Worker+"/"+t.ID, "Task "+t.ID, func() {
					g.clusterOpen(domain.ClusterRoute{Page: poolTaskPage(t), Worker: t.Worker, Pool: f.route.Pool, Target: t.ID})
				})
			}
		}
	}
}

func poolTaskPage(t domain.PoolTaskRow) string {
	if t.Inference {
		return "inference-task"
	}
	return "task"
}
