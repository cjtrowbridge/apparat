package hud

import "fmt"

type Shell struct {
	snapshot  Snapshot
	cancelled bool
}

func NewShell() Shell {
	return NewShellWithConfig(DefaultConfigManager{})
}

func NewShellWithConfig(manager ConfigManager) Shell {
	config := manager.Config()
	tabs := DefaultTabs(config)
	activeIndex := tabIndexByID(tabs, config.TabView.DefaultTab)
	return Shell{snapshot: Snapshot{Config: config, Tabs: tabs, ActiveIndex: activeIndex, FocusIndex: 0, VoiceState: VoiceIdle, Diagnostics: Diagnostics{FrameTime: "mock 16.6ms", Focused: "tab:comrades row:0", ActiveRoute: "mock://hud/comrades", Input: "idle", EventQueueSize: 0, Layout: "1280x800 steam-deck-readable"}}}
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

func (shell *Shell) SelectTabByID(id TabID) error {
	for index, tab := range shell.snapshot.Tabs {
		if tab.ID() == id {
			return shell.SelectTab(index)
		}
	}
	return fmt.Errorf("unknown tab %q", id)
}

func (shell *Shell) MoveFocus(delta int) {
	rows := len(shell.snapshot.ActiveTab().Rows())
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
	shell.snapshot.Diagnostics.Focused = fmt.Sprintf("tab:%s row:%d", active.ID(), shell.snapshot.FocusIndex)
	shell.snapshot.Diagnostics.ActiveRoute = fmt.Sprintf("mock://hud/%s", active.ID())
	shell.snapshot.Diagnostics.Input = input
}
