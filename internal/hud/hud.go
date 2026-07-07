package hud

import "fmt"

type TabID string

const (
	TabComrades TabID = "comrades"
	TabProjects TabID = "projects"
	TabResearch TabID = "research"
	TabCluster  TabID = "cluster"
	TabRouting  TabID = "routing"
	TabTasks    TabID = "tasks"
	TabSettings TabID = "settings"
)

type VoiceState string

const (
	VoiceIdle         VoiceState = "idle"
	VoiceRecording    VoiceState = "recording"
	VoiceQueued       VoiceState = "queued"
	VoiceTranscribing VoiceState = "transcribing"
	VoiceFailed       VoiceState = "failed"
	VoiceComplete     VoiceState = "complete"
)

type Tab struct {
	ID      TabID
	Title   string
	Summary string
	Rows    []string
}

type Diagnostics struct {
	FrameTime      string
	Focused        string
	ActiveRoute    string
	Input          string
	EventQueueSize int
	Layout         string
}

type Snapshot struct {
	Tabs        []Tab
	ActiveIndex int
	FocusIndex  int
	VoiceState  VoiceState
	Diagnostics Diagnostics
}

func (snapshot Snapshot) ActiveTab() Tab {
	return snapshot.Tabs[snapshot.ActiveIndex]
}

func (snapshot Snapshot) TabTitles() []string {
	titles := make([]string, 0, len(snapshot.Tabs))
	for _, tab := range snapshot.Tabs {
		titles = append(titles, tab.Title)
	}
	return titles
}

type Shell struct {
	snapshot  Snapshot
	cancelled bool
}

func NewShell() Shell {
	tabs := []Tab{
		{ID: TabComrades, Title: "Comrades", Summary: "Future chat and shared compute grants.", Rows: []string{"Placeholder: trusted friends", "Placeholder: comrade queues", "Default: lower priority than owner work"}},
		{ID: TabProjects, Title: "Projects", Summary: "Mock project chats, files, artifacts, and Git state.", Rows: []string{"apparat/ clean", "chat: architecture sketch", "artifact: mock transcript"}},
		{ID: TabResearch, Title: "Research", Summary: "Future BOINC and validated research compute.", Rows: []string{"Placeholder: validated projects", "Budget: opt-in only", "Priority: below personal work"}},
		{ID: TabCluster, Title: "Cluster", Summary: "Mock device health and typed capabilities.", Rows: []string{"steamdeck: GUI console", "worker: text_generation, speech_to_text", "workstation: image_generation, video_generation, research_boinc"}},
		{ID: TabRouting, Title: "Routing", Summary: "Mock queues, compatibility filters, and fallbacks.", Rows: []string{"text queue: OpenAI-compatible -> Ollama -> llama.cpp", "speech queue: whisper.cpp", "research queue: BOINC overnight"}},
		{ID: TabTasks, Title: "Tasks", Summary: "Mock schedules, webhooks, events, approvals, and run history.", Rows: []string{"daily summary: scheduled", "webhook: pending approval", "event rule: route failed job"}},
		{ID: TabSettings, Title: "Settings", Summary: "Mock identity, networking, database, audio, and diagnostics.", Rows: []string{"identity: local device", "network: WireGuard/LAN HTTPS", "audio: R2/right-Ctrl push-to-talk"}},
	}
	return Shell{
		snapshot: Snapshot{
			Tabs:        tabs,
			ActiveIndex: 0,
			FocusIndex:  0,
			VoiceState:  VoiceIdle,
			Diagnostics: Diagnostics{
				FrameTime:      "mock 16.6ms",
				Focused:        "tab:comrades row:0",
				ActiveRoute:    "mock://hud/comrades",
				Input:          "idle",
				EventQueueSize: 0,
				Layout:         "1280x800 steam-deck-readable",
			},
		},
	}
}

func (shell *Shell) Snapshot() Snapshot {
	return shell.snapshot
}

func (shell *Shell) NextTab() {
	shell.snapshot.ActiveIndex = (shell.snapshot.ActiveIndex + 1) % len(shell.snapshot.Tabs)
	shell.snapshot.FocusIndex = 0
	shell.updateDiagnostics("next-tab")
}

func (shell *Shell) PreviousTab() {
	shell.snapshot.ActiveIndex = (shell.snapshot.ActiveIndex + len(shell.snapshot.Tabs) - 1) % len(shell.snapshot.Tabs)
	shell.snapshot.FocusIndex = 0
	shell.updateDiagnostics("previous-tab")
}

func (shell *Shell) SelectTab(index int) error {
	if index < 0 || index >= len(shell.snapshot.Tabs) {
		return fmt.Errorf("tab index %d out of range", index)
	}
	shell.snapshot.ActiveIndex = index
	shell.snapshot.FocusIndex = 0
	shell.updateDiagnostics("select-tab")
	return nil
}

func (shell *Shell) MoveFocus(delta int) {
	rows := len(shell.snapshot.ActiveTab().Rows)
	if rows == 0 {
		shell.snapshot.FocusIndex = 0
		return
	}
	shell.snapshot.FocusIndex = (shell.snapshot.FocusIndex + delta + rows) % rows
	shell.updateDiagnostics("move-focus")
}

func (shell *Shell) StartVoiceCapture(input string) {
	shell.cancelled = false
	shell.snapshot.VoiceState = VoiceRecording
	shell.snapshot.Diagnostics.Input = input
}

func (shell *Shell) CancelVoiceCapture() {
	if shell.snapshot.VoiceState == VoiceRecording {
		shell.cancelled = true
		shell.snapshot.VoiceState = VoiceIdle
		shell.snapshot.Diagnostics.Input = "voice-cancelled"
	}
}

func (shell *Shell) ReleaseVoiceCapture() {
	if shell.cancelled {
		shell.cancelled = false
		shell.snapshot.VoiceState = VoiceIdle
		return
	}
	if shell.snapshot.VoiceState == VoiceRecording {
		shell.snapshot.VoiceState = VoiceQueued
		shell.snapshot.Diagnostics.Input = "voice-submitted"
	}
}

func (shell *Shell) updateDiagnostics(input string) {
	active := shell.snapshot.ActiveTab()
	shell.snapshot.Diagnostics.Focused = fmt.Sprintf("tab:%s row:%d", active.ID, shell.snapshot.FocusIndex)
	shell.snapshot.Diagnostics.ActiveRoute = fmt.Sprintf("mock://hud/%s", active.ID)
	shell.snapshot.Diagnostics.Input = input
}
