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
