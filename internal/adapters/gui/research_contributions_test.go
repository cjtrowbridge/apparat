//go:build gui

package gui

import (
	"testing"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/event"
)

func TestResearchContributionHeadersResortMockFriends(t *testing.T) {
	game := NewGame()
	section := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[3].Sections[1]
	content := game.buildSectionContainer(section)
	content.Validate()
	assertContributionOrder(t, collectTextLabels(content), []string{"Mara", "River", "Zvyo", "Puck"})
	if !containsLabel(collectTextLabels(content), "Your Contribution: 2.4 pflops mock total") {
		t.Fatal("personal contribution was not rendered separately")
	}
	if header := findButtonByLabel(content, "Friend"); header == nil {
		t.Fatal("friend sort header was not rendered")
	} else {
		header.Click()
		event.ExecuteDeferred()
	}
	content = game.buildSectionContainer(section)
	content.Validate()
	assertContributionOrder(t, collectTextLabels(content), []string{"Mara", "Puck", "River", "Zvyo"})
	if header := findButtonByLabel(content, "Contribution"); header == nil {
		t.Fatal("contribution sort header was not rendered")
	} else {
		header.Click()
		event.ExecuteDeferred()
	}
	content = game.buildSectionContainer(section)
	content.Validate()
	assertContributionOrder(t, collectTextLabels(content), []string{"Puck", "River", "Zvyo", "Mara"})
}

func assertContributionOrder(t *testing.T, labels, want []string) {
	t.Helper()
	position := 0
	for _, label := range labels {
		if position < len(want) && label == want[position] {
			position++
		}
	}
	if position != len(want) {
		t.Fatalf("friend order in %#v did not contain %#v", labels, want)
	}
}
