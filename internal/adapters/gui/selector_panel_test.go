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
	heading := game.selectorHeading(cluster.Sections[0])
	texts := collectTextNodes(heading)
	if len(texts) != 2 || texts[0].Label != "DEVICES" || texts[1].Label != cluster.Sections[0].Description {
		t.Fatalf("selector heading texts = %#v, want title and description", texts)
	}
	for _, text := range texts {
		if text.MaxWidth <= 0 {
			t.Fatalf("selector heading text %q MaxWidth = %.1f, want positive", text.Label, text.MaxWidth)
		}
	}
	game.selectSection(cluster.ID(), 5)
	details := game.detailSections(cluster)
	if len(details) != 1 || details[0].Title != "Tasks" {
		t.Fatalf("selected cluster details = %#v, want Tasks", details)
	}
}
