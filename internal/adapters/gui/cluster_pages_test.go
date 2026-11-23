//go:build gui

// Checks that the Cluster gear and Back share the responsive detail stack.
// Covers navigation with an empty cluster and retained parent scroll.
package gui

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/widget"
	"slices"
	"testing"
	"time"
)

func TestClusterSettingsNavigation(t *testing.T) {
	for _, width := range []int{460, 1200} {
		g := NewGame()
		g.width = width
		selectTabThroughStrip(t, g, hud.TabCluster)
		g.openClusterSettings()
		g.rebuildUI(g.shell.Snapshot())
		parent := g.clusterTop()
		if parent == nil || parent.page.Title != "Cluster settings" || g.cluster.gear == nil || !g.detailOpen[hud.TabCluster] {
			t.Fatalf("settings unavailable at width %d", width)
		}
		g.cluster.pane.ScrollTop = .4
		g.clusterPush(clusterFrame{page: domain.ClusterPage{Title: "Pools"}})
		g.clusterBack()
		if g.clusterTop().scroll != .4 {
			t.Fatal("lost parent scroll")
		}
		g.clusterBack()
		if g.clusterTop() != nil || g.detailOpen[hud.TabCluster] {
			t.Fatal("back did not restore list")
		}
	}
}

func TestDeviceDirectoryPoolNavigation(t *testing.T) {
	g := NewGame()
	roster := []domain.PoolDevice{{ID: "d", Host: "Phone", OS: "android", Pools: []string{"Chat Pool"}}, {ID: "w", Worker: true, Pools: []string{"Chat Pool"}}}
	g.serviceProjection.Devices = []domain.ClusterDevice{{Fingerprint: "w", Host: "Worker", Admin: true, Roster: roster}}
	g.serviceProjection.LocalDevice = roster[0]
	g.serviceProjection.Peers = []domain.PoolDevice{{ID: "d", Host: "localhost"}, {ID: "other", Host: "Phone"}}
	rows := g.personalSections()
	if len(rows) != 2 || rows[0].SelectorIcon != "1f4f1" || !rows[0].CurrentDevice || rows[1].CurrentDevice || rows[0].Title != "Phone" {
		t.Fatal(rows)
	}
	opts := g.deviceManagementLinks(rows[0].ActionID)
	if len(opts) != 2 || opts[1].Route.Page != "performance" || opts[0].Label != "Rename Phone" || opts[0].Route.Page != "rename-local" || opts[0].Route.Target != "d" || opts[0].Route.Worker != "w" {
		t.Fatal(opts)
	}
	section := g.routingMembers(hud.Section{ActionID: "chat-pool"})
	if len(section.Rows) != 2 || section.Rows[1].Detail != "Worker" {
		t.Fatal(section)
	}
	g.serviceProjection.Devices[0].Admin = false
	opts = g.deviceManagementLinks("gui-d")
	if len(opts) != 2 || opts[1].Route.Page != "performance" || opts[0].Route.Worker != "" || opts[0].Route.Page != "rename-local" || len(g.deviceManagementLinks("worker-w")) != 0 {
		t.Fatal("nonadmin pool editor")
	}
}

func TestPoolSelectionSaveAndCancel(t *testing.T) {
	for _, button := range []string{"Cancel", "Save changes"} {
		g := NewGame()
		saved := make(chan []string, 1)
		page := domain.ClusterPage{Pools: []string{"Chat", "Research"}, Assigned: []string{"Chat"}, SavePools: func(_ context.Context, p []string) error {
			saved <- p
			return nil
		}}
		selectTabThroughStrip(t, g, hud.TabCluster)
		g.clusterPush(clusterFrame{route: domain.ClusterRoute{Page: "device-pools"}, page: page, selectedPools: map[string]bool{"Chat": true}})
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		var checks []*widget.Checkbox
		for _, w := range g.cluster.pane.GetFocusers() {
			c, ok := w.(*widget.Checkbox)
			if ok {
				checks = append(checks, c)
			}
		}
		if len(checks) != 2 || checks[0].State() != widget.WidgetChecked || checks[1].State() == widget.WidgetChecked {
			t.Fatal("expected saved pool checks")
		}
		iw, _ := g.theme.CheckboxTheme.Image.Unchecked.MinSize()
		if iw < 16 {
			t.Fatal("invisible checkbox")
		}
		checks[0].SetState(widget.WidgetUnchecked)
		checks[1].SetState(widget.WidgetChecked)
		event.ExecuteDeferred()
		if len(saved) != 0 || g.cluster.pending != nil {
			t.Fatal("checkbox submitted a change")
		}
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		var action *widget.Button
		for _, w := range g.cluster.pane.GetFocusers() {
			b, ok := w.(*widget.Button)
			if ok && b.Text() != nil && b.Text().Label == button {
				action = b
			}
		}
		if action == nil {
			t.Fatal("missing action", button)
		}
		action.Click()
		event.ExecuteDeferred()
		if button == "Cancel" {
			if g.clusterTop() != nil || len(saved) != 0 {
				t.Fatal("cancel submitted or retained draft")
			}
		} else {
			select {
			case p := <-saved:
				if !slices.Equal(p, []string{"Research"}) {
					t.Fatal("wrong saved selection", p)
				}
			case <-time.After(time.Second):
				t.Fatal("save did not submit")
			}
		}
	}
}

func TestPoolTabsAndDirectoryReturn(t *testing.T) {
	for _, width := range []int{360, 1200} {
		g := NewGame()
		g.width = width
		g.runtimeInfo.Queries = &projectionSource{}
		selectTabThroughStrip(t, g, hud.TabCluster)
		opt := domain.ClusterOption{Label: "Worker", Route: domain.ClusterRoute{Page: "device-detail", Target: "worker-w"}}
		g.serviceProjection.Devices = []domain.ClusterDevice{{Host: "worker", Fingerprint: "w"}}
		g.clusterPush(clusterFrame{route: domain.ClusterRoute{Page: "pool", Pool: "Chat Pool"}, page: domain.ClusterPage{Title: "Chat Pool", Overview: &domain.PoolOverview{Devices: []domain.ClusterOption{opt}}}})
		f := g.clusterTop()
		if !poolTaskMatches("succeeded", f.taskFilter) || poolTaskMatches("running", f.taskFilter) || f.oldestFirst {
			t.Fatal("unexpected initial task filter or sort order")
		}
		for _, view := range []string{"Tasks", "Authorized users", "Devices"} {
			g.selectPoolView(f, view)
			g.rebuildUI(g.shell.Snapshot())
			renderNavigation(g)
		}
		g.clusterOpen(opt.Route)
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		if g.clusterTop().route.Target != "worker-w" {
			t.Fatal("wrong device")
		}
		g.clusterBack()
		if g.clusterTop().poolTab != "Devices" || g.clusterTop().route.Pool != "Chat Pool" {
			t.Fatal("lost pool destination on Back")
		}
		g.serviceProjection.Peers = []domain.PoolDevice{{Host: "Phone.local."}}
		g.receiveLAN(`["Worker","phone","fresh"]`)
		if !slices.Equal(g.discoverableNames(), []string{"fresh"}) {
			t.Fatal(g.discoverableNames())
		}
		g.openPeer("Worker")
		if g.lan.selected != "" {
			t.Fatal("known worker reopened in discovery")
		}
	}
}
