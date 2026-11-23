//go:build gui

// Owns the Cluster detail navigation stack and the single outstanding worker request.
// Widget rebuilds retain the selected route, scroll position and text draft.
package gui

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
	"slices"
	"time"
)

type clusterSource interface {
	ClusterPage(context.Context, domain.ClusterRoute) (domain.ClusterPage, error)
}
type clusterFrame struct {
	route                       domain.ClusterRoute
	page                        domain.ClusterPage
	form                        *domain.ClusterOption
	draft, notice               string
	selectedPools               map[string]bool
	poolTab, taskFilter         string
	oldestFirst, choosingFilter bool
	scroll                      float64
	focus                       focusBookmark
}
type clusterResponse struct {
	page   domain.ClusterPage
	err    error
	edited bool
}
type clusterNavigation struct {
	stack          []clusterFrame
	pending        chan clusterResponse
	pane           *widget.ScrollContainer
	gear           *widget.Button
	previousDetail bool
}

func (g *Game) clusterTop() *clusterFrame {
	s := g.cluster.stack
	if len(s) == 0 {
		return nil
	}
	return &s[len(s)-1]
}
func (g *Game) clusterRemember() {
	f := g.clusterTop()
	if f != nil {
		if g.cluster.pane != nil {
			f.scroll = g.cluster.pane.ScrollTop
		}
		f.focus = g.nav.memory[hud.TabCluster]
	}
}
func (g *Game) clusterPush(f clusterFrame) {
	if g.cluster.pending != nil {
		return
	}
	if len(g.cluster.stack) == 0 {
		g.cluster.previousDetail = g.detailOpen[hud.TabCluster]
	}
	g.clusterRemember()
	g.cluster.stack = append(g.cluster.stack, f)
	g.detailOpen[hud.TabCluster] = true
	g.layoutDirty = true
}
func (g *Game) clusterOpen(r domain.ClusterRoute) {
	if g.cluster.pending != nil {
		return
	}
	g.clusterPush(clusterFrame{route: r})
	g.clusterLoad()
}
func (g *Game) clusterLoad() {
	f := g.clusterTop()
	if f == nil || g.cluster.pending != nil {
		return
	}
	if f.route.Page == "device-detail" {
		g.layoutDirty = true
		return
	}
	if f.route.Worker == "" && f.route.Page == "settings" {
		f.page = domain.ClusterPage{Title: "Cluster settings"}
		for _, d := range g.serviceProjection.Devices {
			if d.Admin {
				f.page.Options = append(f.page.Options, domain.ClusterOption{Label: d.DisplayName(), Route: domain.ClusterRoute{Worker: d.Fingerprint, Page: "settings"}})
			}
		}
		if len(f.page.Options) == 1 {
			f.route = f.page.Options[0].Route
			g.clusterLoad()
			return
		}
		if len(f.page.Options) == 0 {
			f.page.Detail = "Choose an authenticated worker with administrator access to configure cluster policy."
		}
		g.layoutDirty = true
		return
	}
	source, ok := g.runtimeInfo.Queries.(clusterSource)
	if !ok {
		f.notice = "Cluster settings are unavailable."
		return
	}
	r := f.route
	g.clusterRequest(func(ctx context.Context) clusterResponse {
		p, err := source.ClusterPage(ctx, r)
		return clusterResponse{page: p, err: err}
	})
}
func (g *Game) clusterRequest(run func(context.Context) clusterResponse) {
	ch := make(chan clusterResponse, 1)
	g.cluster.pending = ch
	g.layoutDirty = true
	ctx := g.runContext
	if ctx == nil {
		ctx = context.Background()
	}
	go func() {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		ch <- run(ctx)
	}()
}
func (g *Game) pollCluster() {
	select {
	case response := <-g.cluster.pending:
		g.cluster.pending = nil
		f := g.clusterTop()
		if f == nil {
			return
		}
		switch {
		case response.err != nil:
			f.notice = response.err.Error()
		case response.edited && f.page.SavePools != nil:
			g.clusterBack()
		case response.edited:
			if f.form != nil {
				g.clusterBack()
			}
			g.clusterLoad()
		default:
			f.page = response.page
			if f.page.SavePools != nil {
				f.selectedPools = make(map[string]bool)
				for _, pool := range f.page.Assigned {
					f.selectedPools[pool] = true
				}
			}
			if (f.route.Page == "rename" || f.route.Page == "rename-local") && len(f.page.Options) == 1 {
				f.form = &f.page.Options[0]
			}
			f.notice = ""
		}
		g.layoutDirty = true
	default:
	}
}
func (g *Game) clusterBack() {
	if g.cluster.pending != nil {
		return
	}
	s := g.cluster.stack
	if len(s) == 0 {
		return
	}
	g.cluster.stack = s[:len(s)-1]
	f := g.clusterTop()
	if f != nil {
		if f.route.Page == "pool" || f.route.Page == "pool-policy" && s[len(s)-1].route.Page == "device-pools" {
			g.clusterLoad()
		}
		if g.nav.memory != nil {
			g.nav.memory[hud.TabCluster] = f.focus
		}
	} else {
		g.detailOpen[hud.TabCluster] = g.cluster.previousDetail
		if s[len(s)-1].route.Page == "pool" {
			delete(g.selectedSections, hud.TabCluster)
			g.detailOpen[hud.TabCluster] = false
		}
	}
	g.layoutDirty = true
}
func (g *Game) clusterSelect(opt domain.ClusterOption) {
	if g.cluster.pending != nil {
		return
	}
	if opt.Run == nil {
		g.clusterOpen(opt.Route)
		return
	}
	if opt.Input != "" || opt.Confirm != "" {
		g.clusterPush(clusterFrame{page: domain.ClusterPage{Title: opt.Label, Detail: opt.Confirm}, form: &opt})
		return
	}
	g.clusterRun(opt, "")
}
func (g *Game) clusterRun(opt domain.ClusterOption, value string) {
	if g.cluster.pending != nil {
		return
	}
	g.clusterRequest(func(ctx context.Context) clusterResponse {
		return clusterResponse{edited: true, err: opt.Run(ctx, value)}
	})
}
func (g *Game) hudIcon(name string, click func()) *widget.Button {
	b := widget.NewButton(widget.ButtonOpts.Image(tabButtonImage()), widget.ButtonOpts.Graphic(&widget.GraphicImage{Idle: emojiIcon(name)}), widget.ButtonOpts.GraphicPadding(widget.Insets{}), widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(48, 44)), widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
		if !g.bodySelectionSuppressed() {
			click()
		}
	}))
	b.GetWidget().SetTheme(&widget.Theme{})
	return b
}
func (g *Game) clusterHeader(title *widget.Text) *widget.Container {
	row := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(2), widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{false}))))
	row.GetWidget().LayoutData = widget.GridLayoutData{MaxWidth: g.masterPanePreferredWidth()}
	row.AddChild(title)
	g.cluster.gear = g.hudIcon("2699", g.openClusterSettings)
	row.AddChild(g.cluster.gear)
	return row
}
func (g *Game) openClusterSettings() {
	if g.cluster.pending != nil {
		return
	}
	if len(g.cluster.stack) > 0 {
		g.detailOpen[hud.TabCluster] = g.cluster.previousDetail
		g.cluster.stack = nil
	}
	g.clusterOpen(domain.ClusterRoute{Page: "settings"})
}
func (g *Game) clusterContent(c *widget.Container) {
	f := g.clusterTop()
	c.AddChild(g.hudIcon("1f519", g.clusterBack))
	if f.route.Page == "device-detail" {
		g.poolDeviceDetail(c, f.route.Target)
		return
	}
	c.AddChild(widget.NewText(widget.TextOpts.Text(f.page.Title, g.pageTitleFace, g.theme.ButtonTheme.TextColor.Idle), widget.TextOpts.MaxWidth(float64(max(1, g.detailPanePreferredWidth()-24)))))
	if f.page.Detail != "" {
		c.AddChild(g.detailText(f.page.Detail))
	}
	if f.notice != "" {
		c.AddChild(g.detailText(f.notice))
	}
	if g.cluster.pending != nil {
		c.AddChild(g.detailText("Contacting worker…"))
		return
	}
	if f.page.Overview != nil {
		g.poolTabs(c, f)
		return
	}
	if f.page.SavePools != nil {
		g.poolControls(c, f)
		return
	}
	if f.form != nil {
		if f.form.Input != "" {
			g.projectField(c, f.form.Input, &f.draft)
		}
		g.projectButton(c, f.form.Label, func() {
			g.clusterRun(*f.form, f.draft)
		})
	} else {
		for _, opt := range f.page.Options {
			g.projectButton(c, opt.Label, func() {
				g.clusterSelect(opt)
			})
		}
		if f.route.Page != "" {
			g.projectButton(c, "Refresh", g.clusterLoad)
		}
	}
}
func (g *Game) workerPage(action, page string, admin bool) func() {
	for _, d := range g.serviceProjection.Devices {
		if action == "worker-"+d.Fingerprint && d.ReplicaRevision > 0 && (admin && d.Admin || !admin && d.PoolCount > 0) {
			return func() {
				g.clusterOpen(domain.ClusterRoute{Worker: d.Fingerprint, Page: page})
			}
		}
	}
	return nil
}

func (game *Game) addHeading(shell *widget.Container, title string, id hud.TabID) {
	heading := widget.NewText(
		widget.TextOpts.Text(title, game.pageTitleFace, game.theme.ButtonTheme.TextColor.Idle),
		widget.TextOpts.MaxWidth(float64(game.hudPreferredWidth())),
	)
	if id == hud.TabCluster {
		shell.AddChild(game.clusterHeader(heading))
	} else {
		game.cluster.gear = nil
		shell.AddChild(heading)
	}
}

func (g *Game) poolControls(c *widget.Container, f *clusterFrame) {
	for _, pool := range f.page.Pools {
		label := wrapSelectorLabel(pool, *g.theme.ButtonTheme.TextFace, int(g.detailTextMaxWidth())-36)
		c.AddChild(widget.NewCheckbox(
			widget.CheckboxOpts.Text(label, g.theme.ButtonTheme.TextFace, g.theme.LabelTheme.Color),
			widget.CheckboxOpts.Image(g.theme.CheckboxTheme.Image),
			widget.CheckboxOpts.InitialState(checkedState(f.selectedPools[pool])),
			widget.CheckboxOpts.WidgetOpts(widget.WidgetOpts.MinSize(0, 44)),
			widget.CheckboxOpts.StateChangedHandler(func(a *widget.CheckboxChangedEventArgs) {
				if g.bodySelectionSuppressed() {
					g.layoutDirty = true
					return
				}
				f.selectedPools[pool] = a.State == widget.WidgetChecked
			})))
	}
	if f.notice != "" {
		g.projectButton(c, "Reload saved selection", g.clusterLoad)
	}
	g.projectButton(c, "Save changes", func() {
		g.savePoolSelection(f)
	})
	g.projectButton(c, "Cancel", g.clusterBack)
}
func (g *Game) savePoolSelection(f *clusterFrame) {
	if g.clusterTop() != f || g.cluster.pending != nil {
		return
	}
	selected := []string{}
	for _, pool := range f.page.Pools {
		if f.selectedPools[pool] {
			selected = append(selected, pool)
		}
	}
	if slices.Equal(selected, f.page.Assigned) {
		g.clusterBack()
		return
	}
	save := f.page.SavePools
	g.clusterRequest(func(ctx context.Context) clusterResponse {
		return clusterResponse{edited: true, err: save(ctx, selected)}
	})
}
