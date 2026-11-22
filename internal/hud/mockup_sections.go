// Builds explicitly fictional tab section content.
// Uses package-local contracts and Go standard-library types.
package hud

func comradesMockSections() []Section {
	return append([]Section{
		selectorHeading("NPCs", "Helpful fictional guides for the console.", SelectorPalette[0]),
		mockSection("librarian", "Librarian", "A calm mock conversation about finding the right record.", ContentChat, SelectorPalette[0], "Librarian|I indexed the last cluster note under ‘operations’.", "You|Thanks. What changed?", "Librarian|The routing shelf is ready for review."),
		selectorHeading("Comrades", "Mock trusted people and shared-work conversations.", SelectorPalette[1]),
	}, chatSections(SelectorPalette[1])...)
}

func chatSections(color string) []Section {
	people := []string{"Self", "Zvyo", "Iskra", "Puck", "Kilo", "Glitchi", "Pico", "Neon", "Sprout", "Zephyr", "Bina", "Lumen", "Eco", "Mira", "Veda", "Aura"}
	sections := make([]Section, 0, len(people))
	for _, person := range people {
		sections = append(sections, mockSection(person, person, "A static mock conversation; replies are not sent yet.", ContentChat, color, person+"|Mock check-in: the shared work queue is quiet.", "You|I’m reviewing the next step now.", person+"|Mock acknowledgement recorded locally only."))
	}
	return sections
}

func projectsMockSections() []Section {
	projectColor, pipelineColor := SelectorPalette[0], SelectorPalette[1]
	return []Section{
		selectorHeading("Project Repos", "Browse repositories to explore their files, chats, and actions.", projectColor),
		mockSection("solving-incompleteness", "Solving Incompleteness", "Mock research workspace with Git and Chat views.", ContentProject, projectColor, "Branch|main · mock clean", "Focus|proof notes and formalization", "Chat|mock review waiting for a collaborator"),
		mockSection("invent-quantum-ai", "Invent Quantum AI", "Mock exploratory workspace with Git and Chat views.", ContentProject, projectColor, "Branch|prototype-qai", "Focus|model and hardware notes", "Chat|mock design exchange"),
		mockSection("solve-unification", "Solve Unification", "Mock theory workspace with Git and Chat views.", ContentProject, projectColor, "Branch|geometry-ledger", "Focus|constraint map", "Chat|mock question queue"),
		mockSection("mechanical-computers", "Mechanical Computers", "Mock hardware workspace with Git and Chat views.", ContentProject, projectColor, "Branch|mechanisms", "Focus|timing diagrams", "Chat|mock fabrication review"),
		mockSection("what-s-in-the-news", "What's in the news?", "Mock pipeline workspace with a task entry point.", ContentPipeline, pipelineColor, "Entry point|news_digest", "Input|mock feeds and saved topics", "Run|no-op until task execution exists"),
		mockSection("what-s-happening-tonight", "What's happening tonight?", "Mock pipeline workspace with a task entry point.", ContentPipeline, pipelineColor, "Entry point|tonight_brief", "Input|mock calendar and local events", "Run|no-op until task execution exists"),
		mockSection("oppo-analysis", "Oppo Analysis", "Mock pipeline workspace with a task entry point.", ContentPipeline, pipelineColor, "Entry point|oppo_analysis", "Input|mock notes and route policy", "Run|no-op until task execution exists"),
	}
}

func clusterMockSections() []Section {
	return []Section{
		selectorHeading("Cluster Devices", "Browse nearby Apparat hosts and known devices.", SelectorPalette[0]),
		mockSection("add-new-device", "Add New Device", "Local network discovery", ContentStandard, SelectorPalette[0]),
		mockSection("pixel-2-ai-server", "Pixel 2 AI Server (4gb ram)", "Mock local text and speech worker.", ContentStandard, SelectorPalette[0], "Role|AI server", "Memory|4gb mock capacity", "Status|mock online"),
		mockSection("trash-can-mac-pro", "Trash Can Mac Pro (256gb ram)", "Mock high-memory workstation.", ContentStandard, SelectorPalette[0], "Role|workstation", "Memory|256gb mock capacity", "Status|mock online"),
		selectorHeading("Routing", "How your cluster prioritizes different kinds of tasks.", SelectorPalette[1]),
		mockSection("chat-pool", "Chat Pool (High priority)", "Mock owner conversation routing policy.", ContentStandard, SelectorPalette[1], "Priority|high", "Destinations|mock text-capable devices"),
		mockSection("project-pool", "Project Pool (Medium Priority)", "Mock project workload routing policy.", ContentStandard, SelectorPalette[1], "Priority|medium", "Destinations|mock compatible workers"),
		mockSection("comrade-pool", "Comrade Pool (Low Priority)", "Mock shared-compute routing policy.", ContentStandard, SelectorPalette[1], "Priority|low", "Safety|owner work wins"),
		mockSection("research-pool", "Research Pool (Last Priority)", "Mock donation routing policy.", ContentStandard, SelectorPalette[1], "Priority|last", "Safety|budget gated"),
		selectorHeading("Tasks", "Things your cluster needs to do at certain times or for certain events.", SelectorPalette[2]),
		mockSection("every-hour", "Every Hour", "Mock hourly task schedule.", ContentStandard, SelectorPalette[2], "Entry point|hourly_maintenance", "State|not scheduled"),
		mockSection("every-day", "Every Day", "Mock daily task schedule.", ContentStandard, SelectorPalette[2], "Entry point|daily_brief", "State|not scheduled"),
		mockSection("every-week", "Every Week", "Mock weekly task schedule.", ContentStandard, SelectorPalette[2], "Entry point|weekly_review", "State|not scheduled"),
		mockSection("every-month", "Every Month", "Mock monthly task schedule.", ContentStandard, SelectorPalette[2], "Entry point|monthly_archive", "State|not scheduled"),
		mockSection("webhooks", "Webhooks", "Mock authenticated task triggers.", ContentStandard, SelectorPalette[2], "Endpoint|not configured", "State|disabled"),
		selectorHeading("Inference Types", "Which devices in your cluster are set up for each kind of inference.", SelectorPalette[3]),
		mockSection("text-generation", "Text Generation", "Mock text capability inventory.", ContentStandard, SelectorPalette[3], "Devices|Pixel 2 AI Server", "State|mock available"),
		mockSection("image-generation", "Image Generation", "Mock image capability inventory.", ContentStandard, SelectorPalette[3], "Devices|Trash Can Mac Pro", "State|mock available"),
		mockSection("text-to-speech", "Text-to-Speech", "Mock TTS capability inventory.", ContentStandard, SelectorPalette[3], "Devices|Pixel 2 AI Server", "State|mock available"),
		mockSection("speech-to-text", "Speech-to-Text", "Mock STT capability inventory.", ContentStandard, SelectorPalette[3], "Devices|Pixel 2 AI Server", "State|mock available"),
		mockSection("video-generation", "Video Generation", "Mock video capability inventory.", ContentStandard, SelectorPalette[3], "Devices|Trash Can Mac Pro", "State|mock planned"),
	}
}

func researchMockSections() []Section {
	return []Section{
		selectorHeading("Your Research", "The total compute power you've donated to each type of research.", SelectorPalette[0]),
		researchContributionSection("curing-cancer", "Curing Cancer (2.4 pflops)", "2.4 pflops mock total", "Focus|medical research"),
		researchContributionSection("finding-aliens", "Finding Aliens (6.2 gflops)", "6.2 gflops mock total", "Focus|signal analysis"),
		selectorHeading("Other Research Opportunities", "Mock opportunities awaiting technical validation.", SelectorPalette[1]),
		mockSection("drug-research", "Drug Research", "Mock candidate research opportunity.", ContentStandard, SelectorPalette[1], "Status|not enrolled", "Budget|not assigned"),
		mockSection("einstein-home", "Einstein@Home", "Analyze gravity waves to help find new neutron stars.", ContentStandard, SelectorPalette[1], "Status|not enrolled", "Evidence|mock review pending"),
		mockSection("nfs-home", "NFS@Home", "Find new factorizations of large integers.", ContentStandard, SelectorPalette[1], "Status|not enrolled", "Evidence|mock review pending"),
	}
}

func researchContributionSection(id, title, contribution string, entries ...string) Section {
	return Section{ActionID: id, Title: title, Description: "Mock contribution record; no work is running.", Rows: mockRows(entries...), ContentKind: ContentResearchContribution, SelectorColor: SelectorPalette[0], YourContribution: contribution, FriendContributions: []FriendContribution{{Friend: "Mara", GFlops: 18.4}, {Friend: "River", GFlops: 7.1}, {Friend: "Zvyo", GFlops: 7.1}, {Friend: "Puck", GFlops: 2.8}}}
}

func mockSection(id, title, description string, kind ContentKind, color string, entries ...string) Section {
	return Section{ActionID: id, Title: title, Description: description, Rows: mockRows(entries...), ContentKind: kind, SelectorColor: color}
}

func mockRows(entries ...string) []Row {
	rows := make([]Row, 0, len(entries))
	for _, entry := range entries {
		for index, char := range entry {
			if char == '|' {
				rows = append(rows, Row{Label: entry[:index], Detail: entry[index+1:], Future: true})
			}
		}
	}
	return rows
}
