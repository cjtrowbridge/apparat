package hud

import "testing"

func TestCanonicalTabOrder(t *testing.T) {
	shell := NewShell()
	got := shell.Snapshot().TabTitles()
	want := []string{"Comrades", "Projects", "Research", "Cluster", "Tasks", "Settings"}
	if len(got) != len(want) {
		t.Fatalf("got %d tabs, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("tab %d = %q, want %q", index, got[index], want[index])
		}
	}
}

func TestDefaultHUDConfig(t *testing.T) {
	config := DefaultConfigManager{}.Config()
	if config.Display.Theme != ThemeDark {
		t.Fatalf("theme = %q, want dark", config.Display.Theme)
	}
	if config.TabView.Placement != TabPlacementTop {
		t.Fatalf("placement = %q, want top", config.TabView.Placement)
	}
	if config.Display.Scale != 1.0 || config.Display.FontSize != 18 {
		t.Fatalf("display defaults = scale %v font %d", config.Display.Scale, config.Display.FontSize)
	}
	if config.Privacy.SharingDefault != "disabled" {
		t.Fatalf("sharing default = %q, want disabled", config.Privacy.SharingDefault)
	}
}

func TestTabDescriptorsAreStableAndAccessible(t *testing.T) {
	descriptors := DefaultTabDescriptors()
	ids := []TabID{TabComrades, TabProjects, TabResearch, TabCluster, TabTasks, TabSettings}
	for index, id := range ids {
		descriptor := descriptors[index]
		if descriptor.ID != id {
			t.Fatalf("descriptor %d = %q, want %q", index, descriptor.ID, id)
		}
		if descriptor.Label == "" || descriptor.AccessibilityLabel == "" || !descriptor.Visible || !descriptor.Enabled {
			t.Fatalf("descriptor missing required metadata: %+v", descriptor)
		}
	}
}

func TestTabWrapBehavior(t *testing.T) {
	shell := NewShell()
	shell.PreviousTab()
	if shell.Snapshot().ActiveTab().ID() != TabSettings {
		t.Fatalf("previous from first tab = %q, want settings", shell.Snapshot().ActiveTab().ID())
	}
	shell.NextTab()
	if shell.Snapshot().ActiveTab().ID() != TabComrades {
		t.Fatalf("next from last tab = %q, want comrades", shell.Snapshot().ActiveTab().ID())
	}
}

func TestDirectTabSelection(t *testing.T) {
	shell := NewShell()
	if err := shell.SelectTab(2); err != nil {
		t.Fatal(err)
	}
	if shell.Snapshot().ActiveTab().ID() != TabResearch {
		t.Fatalf("selected tab = %q, want research", shell.Snapshot().ActiveTab().ID())
	}
}

func TestActionRoutingUsesNamedActions(t *testing.T) {
	shell := NewShell()
	if err := shell.ApplyAction(ActionNextTab); err != nil {
		t.Fatal(err)
	}
	if shell.Snapshot().ActiveTab().ID() != TabProjects {
		t.Fatalf("action next selected %q", shell.Snapshot().ActiveTab().ID())
	}
	if err := shell.ApplyAction(ActionPreviousTab); err != nil {
		t.Fatal(err)
	}
	if shell.Snapshot().ActiveTab().ID() != TabComrades {
		t.Fatalf("action previous selected %q", shell.Snapshot().ActiveTab().ID())
	}
}

func TestDefaultBindingsExposeFutureEditableInputs(t *testing.T) {
	bindings := DefaultBindings()
	if len(bindings.Inputs(ActionNextTab)) < 2 {
		t.Fatalf("next-tab bindings too small: %+v", bindings.Inputs(ActionNextTab))
	}
	if len(bindings.Inputs(ActionPushToTalk)) < 2 {
		t.Fatalf("push-to-talk bindings too small: %+v", bindings.Inputs(ActionPushToTalk))
	}
	if len(bindings.Inputs(ActionScroll)) < 4 {
		t.Fatalf("scroll bindings should expose wheel, pointer drag, touch drag, and controller defaults: %+v", bindings.Inputs(ActionScroll))
	}
	if len(bindings.Inputs(ActionScrollUp)) == 0 || len(bindings.Inputs(ActionScrollDown)) == 0 {
		t.Fatalf("scroll step bindings missing: up=%+v down=%+v", bindings.Inputs(ActionScrollUp), bindings.Inputs(ActionScrollDown))
	}
}

func TestRightCtrlCancellationDoesNotSubmitOnRelease(t *testing.T) {
	shell := NewShell()
	shell.StartVoiceCapture("right-ctrl")
	shell.CancelVoiceCapture()
	shell.ReleaseVoiceCapture()
	if shell.Snapshot().VoiceState != VoiceIdle {
		t.Fatalf("voice state = %q, want idle", shell.Snapshot().VoiceState)
	}
}

func TestReleaseVoiceCaptureQueuesSubmission(t *testing.T) {
	shell := NewShell()
	shell.StartVoiceCapture("r2")
	shell.ReleaseVoiceCapture()
	if shell.Snapshot().VoiceState != VoiceQueued {
		t.Fatalf("voice state = %q, want queued", shell.Snapshot().VoiceState)
	}
}

func TestEachTabHasContentAndBackendActionsDisabled(t *testing.T) {
	shell := NewShell()
	for _, tab := range shell.Snapshot().Tabs {
		if tab.Summary == "" || len(tab.Rows()) == 0 {
			t.Fatalf("tab missing content: %+v", tab.Descriptor)
		}
	}
	if !hasDisabledFutureRow(shell.Snapshot().Tabs, TabComrades) {
		t.Fatal("comrades tab should contain disabled future controls")
	}
	if !hasDisabledFutureRow(shell.Snapshot().Tabs, TabTasks) {
		t.Fatal("tasks tab should contain disabled future controls")
	}
}

func TestClusterRoutingAndProjectPipelinesAreGroupedDetailSelectors(t *testing.T) {
	tabs := DefaultTabs(DefaultConfigManager{}.Config())
	cluster := tabByID(t, tabs, TabCluster)
	routing := sectionByTitle(t, cluster.Sections, "Routing")
	if len(routing.DetailSections) < 8 {
		t.Fatalf("routing detail sections = %d, want grouped former routing content", len(routing.DetailSections))
	}
	projects := tabByID(t, tabs, TabProjects)
	pipelines := sectionByTitle(t, projects.Sections, "Pipelines")
	if len(pipelines.DetailSections) != 3 {
		t.Fatalf("pipeline detail sections = %d, want 3", len(pipelines.DetailSections))
	}
	if !hasDisabledFutureRow([]Tab{projects}, TabProjects) {
		t.Fatal("pipelines should keep backend-dependent controls disabled")
	}
}

func tabByID(t *testing.T, tabs []Tab, id TabID) Tab {
	t.Helper()
	for _, tab := range tabs {
		if tab.ID() == id {
			return tab
		}
	}
	t.Fatalf("missing tab %q", id)
	return Tab{}
}

func sectionByTitle(t *testing.T, sections []Section, title string) Section {
	t.Helper()
	for _, section := range sections {
		if section.Title == title {
			return section
		}
	}
	t.Fatalf("missing section %q", title)
	return Section{}
}

func hasDisabledFutureRow(tabs []Tab, id TabID) bool {
	for _, tab := range tabs {
		if tab.ID() != id {
			continue
		}
		for _, section := range tab.Sections {
			for _, row := range section.Rows {
				if row.Disabled && row.Future {
					return true
				}
			}
			for _, detail := range section.DetailSections {
				for _, row := range detail.Rows {
					if row.Disabled && row.Future {
						return true
					}
				}
			}
		}
	}
	return false
}
