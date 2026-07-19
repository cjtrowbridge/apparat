package hud

import "fmt"

type TabID string

const (
	TabComrades TabID = "comrades"
	TabProjects TabID = "projects"
	TabResearch TabID = "research"
	TabCluster  TabID = "cluster"
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

type Row struct {
	Label    string
	Detail   string
	Disabled bool
	Future   bool
}

type Section struct {
	Title          string
	Description    string
	Rows           []Row
	DetailSections []Section
}

type Tab struct {
	Descriptor TabDescriptor
	Summary    string
	Sections   []Section
}

func (tab Tab) ID() TabID     { return tab.Descriptor.ID }
func (tab Tab) Title() string { return tab.Descriptor.Label }

func (tab Tab) Rows() []string {
	rows := []string{}
	for _, section := range tab.Sections {
		rows = append(rows, sectionRows(section)...)
	}
	return rows
}

func sectionRows(section Section) []string {
	rows := make([]string, 0, len(section.Rows))
	for _, row := range section.Rows {
		if row.Detail == "" {
			rows = append(rows, row.Label)
			continue
		}
		rows = append(rows, fmt.Sprintf("%s: %s", row.Label, row.Detail))
	}
	for _, detail := range section.DetailSections {
		rows = append(rows, sectionRows(detail)...)
	}
	return rows
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
	Config      HUDConfig
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
		titles = append(titles, tab.Title())
	}
	return titles
}
