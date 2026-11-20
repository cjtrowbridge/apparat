// Supplies fictional multi-page HUD scenarios.
// Uses package-local contracts and Go standard-library types.
package hud

// scenarioSections provides deliberately fictional data dense enough to exercise
// HUD scrolling before persistence, networking, and service adapters exist.
func scenarioSections(id TabID) []Section {
	switch id {
	case TabComrades:
		return []Section{scenario("conversation-queue", "Conversation Queue", "Mock threads awaiting the future encrypted chat service.", "Mara Chen|mock unread: 3, last message 09:42", "Workshop group|mock planning thread, 12 participants", "River Holt|mock trust review requested", "Family relay|mock offline delivery pending"),
			scenario("trust-and-grants", "Trust And Grants", "Owner-controlled sharing examples; no access has been granted.", "Mara Chen|mock text_generation grant, disabled", "Workshop group|mock weekly quota: 30 minutes", "River Holt|mock block/revoke control", "Audit export|mock 14 recorded grant events"),
			scenario("comrade-queue", "Comrade Queue", "Future low-priority workload visibility.", "Idle capacity|mock policy: owner work first", "Queue depth|mock 4 deferred requests", "Privacy filter|mock project files excluded", "Last decision|mock request declined by schedule"),
			scenario("safety-review", "Safety Review", "Future moderation and recovery surfaces.", "Mutual trust|mock verification expires in 18 days", "Rate limits|mock 6 requests per hour", "Revocation drill|mock not yet run", "Incident notes|mock none retained"),
			scenario("activity-history", "Activity History", "Mock audit events.", "09:42|mock conversation preview indexed", "09:18|mock capacity grant evaluated", "Yesterday|mock trusted group created", "Last week|mock recovery phrase reminder"),
			scenario("planned-controls", "Planned Controls", "Disabled until durable identity and authorization exist.", "Invite comrade|mock disabled", "Create group|mock disabled", "Set queue quota|mock disabled", "Export audit trail|mock disabled")}
	case TabProjects:
		return []Section{scenario("workspace-catalog", "Workspace Catalog", "Mock local and remote project entries.", "apparat|mock active HUD migration", "field-notes|mock offline draft repository", "garden-sensor|mock device telemetry workspace", "research-ledger|mock contribution notebook"),
			scenario("project-chat", "Project Chat", "Future project-scoped conversation timeline.", "Architecture review|mock 18 messages", "Release handoff|mock awaiting owner", "Build investigation|mock linked to artifact", "Offline note|mock unsynced draft"),
			scenario("files-and-drafts", "Files And Drafts", "Filesystem remains authoritative; entries are previews only.", "README.md|mock modified 14 minutes ago", "ROADMAP.md|mock review requested", "notes/scrolling.md|mock local draft", "assets/hud.png|mock generated artifact"),
			scenario("git-activity", "Git Activity", "Future constrained Git operations.", "main|mock clean working tree", "scroll-prototype|mock 2 commits ahead", "Review branch|mock conflict-free", "Stash shelf|mock one protected draft"),
			scenario("artifacts", "Artifacts", "Mock durable references, not real file transfers.", "HUD recording|mock 24 MB preview", "Build report|mock Windows artifact", "Test transcript|mock focused GUI run", "Design sketch|mock tablet annotation"),
			scenario("transactions", "Transactions", "Future owner-device mutation workflow.", "Rename workspace|mock awaiting approval", "Apply patch|mock idempotency key staged", "Publish artifact|mock transfer paused", "Resolve conflict|mock editable failure state")}
	case TabResearch:
		return []Section{scenario("candidate-projects", "Candidate Projects", "Mock candidates awaiting technical and governance review.", "Open climate model|mock CPU profile", "Protein folding study|mock accelerator profile", "City ecology survey|mock storage profile", "Asteroid catalog|mock bandwidth profile"),
			scenario("validation-evidence", "Validation Evidence", "Mock evidence ledger; popularity never substitutes for review.", "Operator identity|mock verification pending", "Source audit|mock reproducible build requested", "Privacy review|mock network policy drafted", "Scientific review|mock external references queued"),
			scenario("resource-budgets", "Resource Budgets", "Owner-authorized example budgets.", "Steam Deck|mock disabled on battery", "Workstation GPU|mock 2 hour overnight window", "Worker CPU|mock 25 percent cap", "Network|mock 2 GB weekly ceiling"),
			scenario("work-unit-history", "Work Unit History", "Mock activity records.", "WU-1042|mock queued, no backend", "WU-1039|mock completed locally", "WU-1034|mock paused for owner work", "WU-1021|mock validation evidence attached"),
			scenario("contribution-impact", "Contribution Impact", "Future explanation and provenance views.", "Compute donated|mock 14.2 device-hours", "Energy estimate|mock 0.8 kWh", "Results accepted|mock 9 provisional", "Community review|mock 2 notes"),
			scenario("research-controls", "Research Controls", "Disabled until validated BOINC integration exists.", "Approve project|mock disabled", "Set thermal limit|mock disabled", "Pause all research|mock disabled", "Export provenance|mock disabled")}
	case TabCluster:
		return []Section{scenario("device-inventory", "Device Inventory", "Mock enrolled and known devices.", "steamdeck|mock GUI console, online", "workstation|mock accelerator host, online", "worker|mock headless queue owner", "phone|mock Android console, sleeping"),
			scenario("capabilities", "Capabilities", "Typed workload inventory examples.", "workstation|mock image_generation, video_generation", "worker|mock text_generation, speech_to_text", "steamdeck|mock GUI, audio capture", "phone|mock touch console, notification relay"),
			scenario("health-and-connectivity", "Health And Connectivity", "Mock diagnostics; no cluster transport is active.", "steamdeck|mock latency 12 ms", "workstation|mock disk 62 percent free", "worker|mock queue heartbeat 09:44", "phone|mock last seen 8 minutes ago"),
			scenario("recent-activity", "Recent Activity", "Mock operations timeline.", "09:44|mock capability refresh", "09:41|mock queue lease returned", "09:32|mock worker restart observed", "Yesterday|mock directory snapshot cached"),
			scenario("queue-ownership", "Queue Ownership", "Authoritative-owner examples.", "personal-text|mock worker owned", "image-lab|mock workstation owned", "research-night|mock owner policy paused", "comrade-low|mock no grants active"),
			scenario("diagnostics", "Diagnostics", "Future repair and export surfaces.", "Connection doctor|mock no transport configured", "Directory signature|mock cached sample", "Log bundle|mock ready for review", "Recovery check|mock not scheduled")}
	case TabSettings:
		return []Section{scenario("display-and-accessibility", "Display And Accessibility", "Mock settings that will become durable user preferences.", "Contrast mode|mock high contrast available", "Text scale|mock 100 percent", "Focus ring|mock controller-first", "Safe-area inset|mock platform measured"),
			scenario("storage-and-backup", "Storage And Backup", "Mock local maintenance surfaces.", "Database path|mock runtime-relative", "Backup schedule|mock weekly reminder", "Artifact retention|mock 30 day policy", "Repair mode|mock dry run only"),
			scenario("audio-and-voice", "Audio And Voice", "Mock routed speech configuration.", "Input device|mock system default", "Push-to-talk|mock hold to record", "ASR route|mock worker fallback", "TTS route|mock disabled"),
			scenario("privacy-and-sharing", "Privacy And Sharing", "Mock owner-controlled defaults.", "Sharing default|mock deny", "Log redaction|mock enabled", "Artifact transfer|mock explicit approval", "Comrade grants|mock none active"),
			scenario("platform-diagnostics", "Platform Diagnostics", "Mock cross-platform checks.", "Desktop headers|mock detected at build", "Android SDK|mock host dependent", "WireGuard|mock externally configured", "Crash reports|mock local only"),
			scenario("recovery-checklist", "Recovery Checklist", "Mock recovery and support steps.", "Export support bundle|mock disabled", "Verify database|mock scheduled", "Rotate recovery key|mock confirmation required", "Read troubleshooting guide|mock available")}
	default:
		return nil
	}
}

func scenario(id, title string, description string, entries ...string) Section {
	rows := make([]Row, 0, len(entries))
	for _, entry := range entries {
		for index, char := range entry {
			if char == '|' {
				rows = append(rows, Row{Label: entry[:index], Detail: entry[index+1:], Future: true})
				break
			}
		}
	}
	return Section{ActionID: id, Title: title, Description: description, Rows: rows}
}
