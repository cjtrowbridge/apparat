//go:build gui

// Checks that the Cluster gear and Back share the responsive detail stack.
// Covers navigation with an empty cluster and retained parent scroll.
package gui

import (
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"testing"
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
	roster := []domain.PoolDevice{{ID: "d", Host: "Phone", OS: "android", Pools: []string{"Chat Pool"}}, {ID: "w", Worker: true}}
	g.serviceProjection.Devices = []domain.ClusterDevice{{Fingerprint: "w", Host: "Worker", Admin: true, Roster: roster}}
	rows := g.personalSections()
	if len(rows) != 1 || rows[0].SelectorIcon != "1f4f1" {
		t.Fatal(rows)
	}
	opts := g.devicePoolLinks(rows[0].ActionID)
	if len(opts) != 1 || opts[0].Route.Target != "d" || opts[0].Route.Worker != "w" {
		t.Fatal(opts)
	}
	section := g.routingMembers(hud.Section{ActionID: "chat-pool"})
	if len(section.Rows) != 2 || section.Rows[1].Detail != "Phone" {
		t.Fatal(section)
	}
	g.serviceProjection.Devices[0].Admin = false
	if len(g.devicePoolLinks("gui-d")) != 0 {
		t.Fatal("nonadmin pool editor")
	}
}
