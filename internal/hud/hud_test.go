package hud

import "testing"

func TestCanonicalTabOrder(t *testing.T) {
	shell := NewShell()
	got := shell.Snapshot().TabTitles()
	want := []string{"Comrades", "Projects", "Research", "Cluster", "Routing", "Tasks", "Settings"}
	if len(got) != len(want) {
		t.Fatalf("got %d tabs, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("tab %d = %q, want %q", index, got[index], want[index])
		}
	}
}

func TestTabWrapBehavior(t *testing.T) {
	shell := NewShell()
	shell.PreviousTab()
	if shell.Snapshot().ActiveTab().ID != TabSettings {
		t.Fatalf("previous from first tab = %q, want settings", shell.Snapshot().ActiveTab().ID)
	}
	shell.NextTab()
	if shell.Snapshot().ActiveTab().ID != TabComrades {
		t.Fatalf("next from last tab = %q, want comrades", shell.Snapshot().ActiveTab().ID)
	}
}

func TestDirectTabSelection(t *testing.T) {
	shell := NewShell()
	if err := shell.SelectTab(2); err != nil {
		t.Fatal(err)
	}
	if shell.Snapshot().ActiveTab().ID != TabResearch {
		t.Fatalf("selected tab = %q, want research", shell.Snapshot().ActiveTab().ID)
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
