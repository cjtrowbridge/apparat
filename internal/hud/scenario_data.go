package hud

// scenarioSections provides deliberately fictional data dense enough to exercise
// HUD scrolling before persistence, networking, and service adapters exist.
func scenarioSections(id TabID) []Section {
	switch id {
	case TabComrades:
		return []Section{scenario("Conversation Queue", "Mock threads awaiting the future encrypted chat service.", "Mara Chen|mock unread: 3, last message 09:42", "Workshop group|mock planning thread, 12 participants", "River Holt|mock trust review requested", "Family relay|mock offline delivery pending"),
			scenario("Trust And Grants", "Owner-controlled sharing examples; no access has been granted.", "Mara Chen|mock text_generation grant, disabled", "Workshop group|mock weekly quota: 30 minutes", "River Holt|mock block/revoke control", "Audit export|mock 14 recorded grant events"),
			scenario("Comrade Queue", "Future low-priority workload visibility.", "Idle capacity|mock policy: owner work first", "Queue depth|mock 4 deferred requests", "Privacy filter|mock project files excluded", "Last decision|mock request declined by schedule"),
			scenario("Safety Review", "Future moderation and recovery surfaces.", "Mutual trust|mock verification expires in 18 days", "Rate limits|mock 6 requests per hour", "Revocation drill|mock not yet run", "Incident notes|mock none retained"),
			scenario("Activity History", "Mock audit events.", "09:42|mock conversation preview indexed", "09:18|mock capacity grant evaluated", "Yesterday|mock trusted group created", "Last week|mock recovery phrase reminder"),
			scenario("Planned Controls", "Disabled until durable identity and authorization exist.", "Invite comrade|mock disabled", "Create group|mock disabled", "Set queue quota|mock disabled", "Export audit trail|mock disabled")}
	case TabProjects:
		return []Section{scenario("Workspace Catalog", "Mock local and remote project entries.", "apparat|mock active HUD migration", "field-notes|mock offline draft repository", "garden-sensor|mock device telemetry workspace", "research-ledger|mock contribution notebook"),
			scenario("Project Chat", "Future project-scoped conversation timeline.", "Architecture review|mock 18 messages", "Release handoff|mock awaiting owner", "Build investigation|mock linked to artifact", "Offline note|mock unsynced draft"),
			scenario("Files And Drafts", "Filesystem remains authoritative; entries are previews only.", "README.md|mock modified 14 minutes ago", "ROADMAP.md|mock review requested", "notes/scrolling.md|mock local draft", "assets/hud.png|mock generated artifact"),
			scenario("Git Activity", "Future constrained Git operations.", "main|mock clean working tree", "scroll-prototype|mock 2 commits ahead", "Review branch|mock conflict-free", "Stash shelf|mock one protected draft"),
			scenario("Artifacts", "Mock durable references, not real file transfers.", "HUD recording|mock 24 MB preview", "Build report|mock Windows artifact", "Test transcript|mock focused GUI run", "Design sketch|mock tablet annotation"),
			scenario("Transactions", "Future owner-device mutation workflow.", "Rename workspace|mock awaiting approval", "Apply patch|mock idempotency key staged", "Publish artifact|mock transfer paused", "Resolve conflict|mock editable failure state")}
	case TabResearch:
		return []Section{scenario("Candidate Projects", "Mock candidates awaiting technical and governance review.", "Open climate model|mock CPU profile", "Protein folding study|mock accelerator profile", "City ecology survey|mock storage profile", "Asteroid catalog|mock bandwidth profile"),
			scenario("Validation Evidence", "Mock evidence ledger; popularity never substitutes for review.", "Operator identity|mock verification pending", "Source audit|mock reproducible build requested", "Privacy review|mock network policy drafted", "Scientific review|mock external references queued"),
			scenario("Resource Budgets", "Owner-authorized example budgets.", "Steam Deck|mock disabled on battery", "Workstation GPU|mock 2 hour overnight window", "Worker CPU|mock 25 percent cap", "Network|mock 2 GB weekly ceiling"),
			scenario("Work Unit History", "Mock activity records.", "WU-1042|mock queued, no backend", "WU-1039|mock completed locally", "WU-1034|mock paused for owner work", "WU-1021|mock validation evidence attached"),
			scenario("Contribution Impact", "Future explanation and provenance views.", "Compute donated|mock 14.2 device-hours", "Energy estimate|mock 0.8 kWh", "Results accepted|mock 9 provisional", "Community review|mock 2 notes"),
			scenario("Research Controls", "Disabled until validated BOINC integration exists.", "Approve project|mock disabled", "Set thermal limit|mock disabled", "Pause all research|mock disabled", "Export provenance|mock disabled")}
	case TabCluster:
		return []Section{scenario("Device Inventory", "Mock enrolled and known devices.", "steamdeck|mock GUI console, online", "workstation|mock accelerator host, online", "worker|mock headless queue owner", "phone|mock Android console, sleeping"),
			scenario("Capabilities", "Typed workload inventory examples.", "workstation|mock image_generation, video_generation", "worker|mock text_generation, speech_to_text", "steamdeck|mock GUI, audio capture", "phone|mock touch console, notification relay"),
			scenario("Health And Connectivity", "Mock diagnostics; no cluster transport is active.", "steamdeck|mock latency 12 ms", "workstation|mock disk 62 percent free", "worker|mock queue heartbeat 09:44", "phone|mock last seen 8 minutes ago"),
			scenario("Recent Activity", "Mock operations timeline.", "09:44|mock capability refresh", "09:41|mock queue lease returned", "09:32|mock worker restart observed", "Yesterday|mock directory snapshot cached"),
			scenario("Queue Ownership", "Authoritative-owner examples.", "personal-text|mock worker owned", "image-lab|mock workstation owned", "research-night|mock owner policy paused", "comrade-low|mock no grants active"),
			scenario("Diagnostics", "Future repair and export surfaces.", "Connection doctor|mock no transport configured", "Directory signature|mock cached sample", "Log bundle|mock ready for review", "Recovery check|mock not scheduled")}
	case TabTasks:
		return []Section{scenario("Task Catalog", "Mock durable-task definitions.", "Morning summary|mock schedule 08:00", "Release watcher|mock webhook trigger", "Research window|mock quiet-hours schedule", "Project backup|mock approval required"),
			scenario("Pending Approvals", "Human authorization examples.", "Publish artifact|mock awaiting owner", "Rotate device key|mock requires confirmation", "Run backup|mock storage cost estimate", "Share compute grant|mock policy review"),
			scenario("Run History", "Mock workflow and retry records.", "run-204|mock completed in 14 seconds", "run-203|mock waiting for worker", "run-202|mock retry scheduled", "run-201|mock cancelled safely"),
			scenario("Webhook Inbox", "Future authenticated trigger examples.", "Build completed|mock source CI", "Sensor alert|mock local bridge", "Calendar event|mock deferred adapter", "Issue mention|mock approval policy"),
			scenario("Failure Review", "Mock durable failure context.", "run-199|mock timeout, retry allowed", "run-198|mock permission denied", "run-197|mock route unavailable", "run-196|mock user cancelled"),
			scenario("Task Controls", "Disabled until scheduler ownership exists.", "Create workflow|mock disabled", "Pause scheduler|mock disabled", "Replay run|mock disabled", "Export history|mock disabled")}
	case TabSettings:
		return []Section{scenario("Display And Accessibility", "Mock settings that will become durable user preferences.", "Contrast mode|mock high contrast available", "Text scale|mock 100 percent", "Focus ring|mock controller-first", "Safe-area inset|mock platform measured"),
			scenario("Storage And Backup", "Mock local maintenance surfaces.", "Database path|mock runtime-relative", "Backup schedule|mock weekly reminder", "Artifact retention|mock 30 day policy", "Repair mode|mock dry run only"),
			scenario("Audio And Voice", "Mock routed speech configuration.", "Input device|mock system default", "Push-to-talk|mock hold to record", "ASR route|mock worker fallback", "TTS route|mock disabled"),
			scenario("Privacy And Sharing", "Mock owner-controlled defaults.", "Sharing default|mock deny", "Log redaction|mock enabled", "Artifact transfer|mock explicit approval", "Comrade grants|mock none active"),
			scenario("Platform Diagnostics", "Mock cross-platform checks.", "Desktop headers|mock detected at build", "Android SDK|mock host dependent", "WireGuard|mock externally configured", "Crash reports|mock local only"),
			scenario("Recovery Checklist", "Mock recovery and support steps.", "Export support bundle|mock disabled", "Verify database|mock scheduled", "Rotate recovery key|mock confirmation required", "Read troubleshooting guide|mock available")}
	default:
		return nil
	}
}

func routingScenarioSections() []Section {
	return []Section{scenario("Routing Profiles", "Mock policy profiles for different work types.", "Personal chat|mock text route preferred", "Project review|mock privacy constrained", "Image workshop|mock workstation first", "Voice capture|mock worker ASR fallback"),
		scenario("Compatible Destinations", "Compatibility filters run before priority.", "text_generation|mock worker then workstation", "image_generation|mock workstation only", "speech_to_text|mock worker only", "research_boinc|mock schedule-gated pool"),
		scenario("Queue State", "Mock queue ownership and pressure.", "personal-text|mock depth 2, owner worker", "image-lab|mock depth 0, owner workstation", "voice-asr|mock depth 1, retry ready", "research-night|mock paused by schedule"),
		scenario("Fallback Order", "Ordered examples, not dynamic optimization.", "Project chat|mock worker -> workstation -> local draft", "ASR|mock worker -> offline capture", "Image|mock workstation -> deferred queue", "Research|mock budget hold -> skipped"),
		scenario("Service Health", "Mock adapter diagnostics.", "OpenAI-compatible|mock endpoint unchecked", "Ollama|mock inventory stale", "llama.cpp|mock local adapter planned", "whisper.cpp|mock portable reference"),
		scenario("Policy Events", "Mock routing explanations.", "09:43|mock image request rejected: capability", "09:35|mock text request admitted", "Yesterday|mock fallback profile changed", "Last week|mock privacy rule reviewed")}
}

func scenario(title string, description string, entries ...string) Section {
	rows := make([]Row, 0, len(entries))
	for _, entry := range entries {
		for index, char := range entry {
			if char == '|' {
				rows = append(rows, Row{Label: entry[:index], Detail: entry[index+1:], Future: true})
				break
			}
		}
	}
	return Section{Title: title, Description: description, Rows: rows}
}
