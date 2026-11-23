//go:build gui

// Exercises the live catalog through the command, search and workspace selection paths.
// The rendered browser and chooser share the same navigation targets as other inputs.
package gui

import (
	"image"
	"strings"
	"testing"

	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/widget"
)

func TestGameAndTabSpokenNames(t *testing.T) {
	g := NewGame()
	defer g.Close()
	for _, item := range g.actionCatalog() {
		if item.Address.Kind != "game" && item.Address.Kind != "tab" {
			continue
		}
		match := g.queryActions("open "+item.Name, true)
		if !match.Exact || len(match.Choices) != 1 || match.Choices[0].Address != item.Address {
			t.Fatalf("open %s must resolve uniquely to %v, got %+v", item.Name, item.Address, match)
		}
	}
	for phrase, want := range map[string]hud.ActionAddress{
		"open cluster":       {Operation: "open", Kind: "tab", Target: "cluster"},
		"open bubble volley": {Operation: "open", Kind: "game", Owner: "stmpck", Target: "14"},
	} {
		match := g.queryActions(phrase, true)
		if !match.Exact || len(match.Choices) != 1 || match.Choices[0].Address != want {
			t.Fatalf("%q resolved to %+v, want %v", phrase, match, want)
		}
	}
}

func TestCatalogFollowsRegistrations(t *testing.T) {
	g := NewGame()
	defer g.Close()
	seen := map[hud.ActionAddress]bool{}
	for _, item := range g.actionCatalog() {
		if item.Address.Target == "" || seen[item.Address] {
			t.Fatalf("missing or duplicate address for %s", item.Path)
		}
		seen[item.Address] = true
	}
	g.projects.store = layoutRegistry{}
	for _, tab := range g.shell.Snapshot().Tabs {
		m := g.queryActions("switch to tab "+tab.Title(), true)
		if !m.Exact || len(m.Choices) != 1 || m.Choices[0].Address.Target != string(tab.ID()) {
			t.Fatal(tab.Title(), m)
		}
	}
	g.projects.rows = []domain.Project{{ID: "abc", Name: "Apparat", Available: true, Threads: []domain.WorkspaceChoice{{ID: "Apparat", Label: "Apparat"}}}}
	g.voiceCommand("open apparat")
	if g.actions.mode != "choose" || len(g.actions.choices) != 2 {
		t.Fatal(g.actions)
	}
	g.voiceCommand("the chat")
	w := &g.projects.workspace
	if w.thread != "Apparat" || w.mode != "conversation" || w.id != "abc" {
		t.Fatal("wrong conversation")
	}
	for _, q := range []string{"open project apparat", "launch project Apparat.", "open apparat project"} {
		m := g.queryActions(q, true)
		if !m.Exact || len(m.Choices) != 1 || m.Choices[0].Address.Target != "abc" {
			t.Fatal(q, m)
		}
	}
	g.voiceCommand("open apparat")
	renderNavigation(g)
	g.voiceCommand("the second one")
	if g.actions.mode != "" {
		t.Fatal("ordinal choice did not dismiss chooser")
	}
	g.voiceCommand("show tasks for apparat")
	if w.mode != "tasks" {
		t.Fatal("workspace action failed")
	}
	key := g.queryActions("open project apparat", true).Choices[0].Address
	g.projects.rows[0].Name = "Other"
	renamed := g.queryActions("open project Other", true)
	if len(renamed.Choices) != 1 || renamed.Choices[0].Address != key || g.queryActions("open project Apparat", true).Exact {
		t.Fatal("rename changed identity or left stale target")
	}
	g.projects.rows = nil
	g.invokeAction(key)
	if !strings.Contains(g.actions.prompt, "no longer available") {
		t.Fatal("stale address accepted")
	}
}

func TestActionBrowserAndExclusiveDictation(t *testing.T) {
	for _, width := range []int{360, 1280} {
		g := NewGame()
		g.Layout(width, 800)
		g.voiceCommand("What can I say?")
		renderNavigation(g)
		ts := g.bodyTargets()
		if len(ts) < 3 || g.shell.Snapshot().ActiveTab().ID() != hud.TabSettings {
			t.Fatal("missing action directory")
		}
		f := ts[1].w.(*widget.TextInput)
		g.ui.SetFocusedWidget(f)
		g.finishUtterance(f, "open Projects")
		if f.GetText() != "open Projects" || g.actions.mode != "browse" {
			t.Fatal("search dictation invoked a command")
		}
		f.ChangedEvent.Fire(&widget.TextInputChangedEventArgs{TextInput: f, InputText: f.GetText()})
		event.ExecuteDeferred()
		g.ui.ClearFocus()
		g.voiceCommand("open project Solving Incompletenes")
		if g.actions.mode != "choose" || len(g.actions.choices) != 1 {
			t.Fatal("near match did not require clarification")
		}
		renderNavigation(g)
		rect := g.bodyTargets()[0].w.GetWidget().Rect
		bounds := image.Rect(width/6, 800/6, width-width/6, 800-800/6)
		if bounds != g.choiceBounds() || rect.Empty() || !rect.In(bounds) {
			t.Fatal("choice outside screen", rect)
		}
		root := g.ui.Container.GetWidget()
		tap := func(x, y int) {
			root.MouseButtonPressedEvent.Fire(&widget.WidgetMouseButtonPressedEventArgs{Widget: root, OffsetX: x, OffsetY: y})
			root.MouseButtonReleasedEvent.Fire(&widget.WidgetMouseButtonReleasedEventArgs{Widget: root, OffsetX: x, OffsetY: y, Inside: true})
			event.ExecuteDeferred()
		}
		tap(width/2, 400)
		if g.actions.mode != "choose" {
			t.Fatal("inside tap dismissed clarification")
		}
		tap(2, 2)
		if g.actions.mode != "browse" || g.actions.query != "open Projects" {
			t.Fatalf("outside tap returned to mode %q, query %q", g.actions.mode, g.actions.query)
		}
		g.voiceCommand("open nonexistentxyz")
		if !strings.Contains(g.actions.prompt, "nonexistentxyz") {
			t.Fatal("missing transcript feedback")
		}
		g.voiceCommand("switch to Research")
		if g.shell.Snapshot().ActiveTab().ID() != hud.TabResearch {
			t.Fatal("could not retry")
		}
		g.Close()
	}
}

func TestVoiceChoiceKeepsProjectContext(t *testing.T) {
	g := NewGame()
	defer g.Close()
	g.projects.store = layoutRegistry{}
	for _, name := range []string{"Alpha", "Beta"} {
		g.projects.rows = append(g.projects.rows, domain.Project{ID: name, Name: name, Available: true, Threads: []domain.WorkspaceChoice{{ID: "Release", Label: "Release"}}})
	}
	var spoken string
	g.runtimeInfo.SpeakChoice = func(s string) {
		spoken = s
	}
	g.prefs.SpokenChoices = true
	g.voiceCommand("open chat Release")
	if len(g.actions.choices) != 2 || !strings.Contains(spoken, "Alpha / Chat") || !strings.Contains(spoken, "Beta / Chat") {
		t.Fatal("missing distinct spoken paths")
	}
	original := g.shell.Snapshot().ActiveTab().ID()
	g.voiceCommand("the chat")
	if g.actions.mode != "choose" || g.shell.Snapshot().ActiveTab().ID() != original {
		t.Fatal("ambiguous follow-up navigated")
	}
	g.projects.rows = g.projects.rows[:1]
	g.voiceCommand("second")
	if g.actions.mode != "choose" || !strings.Contains(g.actions.prompt, "no longer available") {
		t.Fatal("removed thread was invoked")
	}
	g.voiceCommand("cancel")
	if g.actions.mode != "" || spoken != "" {
		t.Fatal("cancel did not dismiss and stop feedback")
	}
}

func TestDiscoveredHostNavigationExpires(t *testing.T) {
	g := NewGame()
	defer g.Close()
	g.receiveLAN(lanNames([]string{"trashcan0", "bad\nname"}))
	match := g.queryActions("open trashcan0", true)
	if !match.Exact || len(match.Choices) != 1 || len(g.lan.names) != 1 {
		t.Fatal(match, g.lan.names)
	}
	id := match.Choices[0].Address
	g.invokeAction(id)
	if g.lan.selected != "trashcan0" || g.shell.Snapshot().ActiveTab().ID() != hud.TabCluster {
		t.Fatal("host detail did not open")
	}
	g.leavePeer()
	g.openPeer("trashcan0")
	g.receiveLAN(lanNames(nil))
	g.invokeAction(id)
	if g.lan.selected != "" || len(g.queryActions("open trashcan0", true).Choices) != 0 {
		t.Fatal("departed peer remained available")
	}
}

func TestPersonalClusterDialogsFollowAvailability(t *testing.T) {
	g := NewGame()
	defer g.Close()
	g.runtimeInfo.EditName = func() {}
	g.runtimeInfo.PairPersonal = func(string, string) {}
	g.runtimeInfo.ClaimHost = func(string, string) {}
	g.receiveLAN(lanNames([]string{"fixture"}))
	if len(g.queryActions("open receive my identity from fixture", true).Choices) != 1 {
		t.Fatal("missing peer pairing action")
	}
	if len(g.queryActions("open claim fixture", true).Choices) != 1 || len(g.queryActions("open edit screen name", true).Choices) != 1 {
		t.Fatal("missing identity or claim action")
	}
	g.receiveLAN(lanNames(nil))
	for _, action := range g.actionCatalog() {
		if action.Address.Kind == "claim-dialog" || action.Address.Kind == "personal-pair-receive" {
			t.Fatal("departed host has claim action")
		}
	}
}
