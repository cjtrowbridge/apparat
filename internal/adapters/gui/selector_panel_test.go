//go:build gui

package gui

import (
	"testing"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
)

func TestClusterSelectorHeadingsAreTextAndTasksOwnsTheContentPanel(t *testing.T) {
	game := NewGame()
	cluster := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[3]
	game.selectSection(cluster.ID(), 0)
	if got := game.selectedSectionIndex(cluster); got != 1 {
		t.Fatalf("heading selection resolved to %d, want first selectable section 1", got)
	}
	body := game.buildMasterDetailTab(cluster)
	if _, ok := unwrapBounded(body).(*widget.Container); !ok {
		t.Fatalf("cluster body type = %T, want *widget.Container", body)
	}
	if heading := game.selectorHeading("Devices"); heading.Label != "DEVICES" || heading.MaxWidth <= 0 {
		t.Fatalf("selector heading = %#v, want visible bounded text", heading)
	}
	game.selectSection(cluster.ID(), 5)
	details := game.detailSections(cluster)
	if len(details) != 1 || details[0].Title != "Tasks" {
		t.Fatalf("selected cluster details = %#v, want Tasks", details)
	}
}
