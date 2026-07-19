package hud

func comradesMockSections() []Section {
	return append([]Section{
		selectorHeading("NPCs", "Helpful fictional guides for the console.", SelectorPalette[0]),
		mockSection("Librarian", "A calm mock conversation about finding the right record.", ContentChat, SelectorPalette[0], "Librarian|I indexed the last cluster note under ‘operations’.", "You|Thanks. What changed?", "Librarian|The routing shelf is ready for review."),
		selectorHeading("Comrades", "Mock trusted people and shared-work conversations.", SelectorPalette[1]),
	}, chatSections(SelectorPalette[1])...)
}

func chatSections(color string) []Section {
	people := []string{"Self", "Zvyo", "Iskra", "Puck", "Kilo", "Glitchi", "Pico", "Neon", "Sprout", "Zephyr", "Bina", "Lumen", "Eco", "Mira", "Veda", "Aura"}
	sections := make([]Section, 0, len(people))
	for _, person := range people {
		sections = append(sections, mockSection(person, "A static mock conversation; replies are not sent yet.", ContentChat, color, person+"|Mock check-in: the shared work queue is quiet.", "You|I’m reviewing the next step now.", person+"|Mock acknowledgement recorded locally only."))
	}
	return sections
}

func projectsMockSections() []Section {
	projectColor, pipelineColor := SelectorPalette[0], SelectorPalette[1]
	return []Section{
		selectorHeading("Projects", "Open-ended workspaces for any kind of work.", projectColor),
		mockSection("Solving Incompleteness", "Mock research workspace with Git and Chat views.", ContentProject, projectColor, "Branch|main · mock clean", "Focus|proof notes and formalization", "Chat|mock review waiting for a collaborator"),
		mockSection("Invent Quantum AI", "Mock exploratory workspace with Git and Chat views.", ContentProject, projectColor, "Branch|prototype-qai", "Focus|model and hardware notes", "Chat|mock design exchange"),
		mockSection("Solve Unification", "Mock theory workspace with Git and Chat views.", ContentProject, projectColor, "Branch|geometry-ledger", "Focus|constraint map", "Chat|mock question queue"),
		mockSection("Mechanical Computers", "Mock hardware workspace with Git and Chat views.", ContentProject, projectColor, "Branch|mechanisms", "Focus|timing diagrams", "Chat|mock fabrication review"),
		selectorHeading("Pipelines", "Project-oriented workspaces with a Cluster-task entry point.", pipelineColor),
		mockSection("What's in the news?", "Mock pipeline workspace with a task entry point.", ContentPipeline, pipelineColor, "Entry point|news_digest", "Input|mock feeds and saved topics", "Run|no-op until task execution exists"),
		mockSection("What's happening tonight?", "Mock pipeline workspace with a task entry point.", ContentPipeline, pipelineColor, "Entry point|tonight_brief", "Input|mock calendar and local events", "Run|no-op until task execution exists"),
		mockSection("Oppo Analysis", "Mock pipeline workspace with a task entry point.", ContentPipeline, pipelineColor, "Entry point|oppo_analysis", "Input|mock notes and route policy", "Run|no-op until task execution exists"),
	}
}

func clusterMockSections() []Section {
	return []Section{
		selectorHeading("Cluster Devices", "Known mock devices and a future enrollment entry point.", SelectorPalette[0]),
		mockSection("Pixel 2 AI Server (4gb ram)", "Mock local text and speech worker.", ContentStandard, SelectorPalette[0], "Role|AI server", "Memory|4gb mock capacity", "Status|mock online"),
		mockSection("Trash Can Mac Pro (256gb ram)", "Mock high-memory workstation.", ContentStandard, SelectorPalette[0], "Role|workstation", "Memory|256gb mock capacity", "Status|mock online"),
		mockSection("(Add New Device)", "Enrollment is not implemented yet.", ContentStandard, SelectorPalette[0], "Enrollment|disabled mock control", "Safety|no device will be added"),
		selectorHeading("Routing", "How your cluster prioritizes different kinds of tasks.", SelectorPalette[1]),
		mockSection("Chat Pool (High priority)", "Mock owner conversation routing policy.", ContentStandard, SelectorPalette[1], "Priority|high", "Destinations|mock text-capable devices"),
		mockSection("Project Pool (Medium Priority)", "Mock project workload routing policy.", ContentStandard, SelectorPalette[1], "Priority|medium", "Destinations|mock compatible workers"),
		mockSection("Comrade Pool (Low Priority)", "Mock shared-compute routing policy.", ContentStandard, SelectorPalette[1], "Priority|low", "Safety|owner work wins"),
		mockSection("Research Pool (Last Priority)", "Mock donation routing policy.", ContentStandard, SelectorPalette[1], "Priority|last", "Safety|budget gated"),
		selectorHeading("Tasks", "Things your cluster needs to do at certain times or for certain events.", SelectorPalette[2]),
		mockSection("Every Hour", "Mock hourly task schedule.", ContentStandard, SelectorPalette[2], "Entry point|hourly_maintenance", "State|not scheduled"),
		mockSection("Every Day", "Mock daily task schedule.", ContentStandard, SelectorPalette[2], "Entry point|daily_brief", "State|not scheduled"),
		mockSection("Every Week", "Mock weekly task schedule.", ContentStandard, SelectorPalette[2], "Entry point|weekly_review", "State|not scheduled"),
		mockSection("Every Month", "Mock monthly task schedule.", ContentStandard, SelectorPalette[2], "Entry point|monthly_archive", "State|not scheduled"),
		mockSection("Webhooks", "Mock authenticated task triggers.", ContentStandard, SelectorPalette[2], "Endpoint|not configured", "State|disabled"),
		selectorHeading("Inference Types", "Which devices in your cluster are set up for each kind of inference.", SelectorPalette[3]),
		mockSection("Text Generation", "Mock text capability inventory.", ContentStandard, SelectorPalette[3], "Devices|Pixel 2 AI Server", "State|mock available"),
		mockSection("Image Generation", "Mock image capability inventory.", ContentStandard, SelectorPalette[3], "Devices|Trash Can Mac Pro", "State|mock available"),
		mockSection("Text-to-Speech", "Mock TTS capability inventory.", ContentStandard, SelectorPalette[3], "Devices|Pixel 2 AI Server", "State|mock available"),
		mockSection("Speech-to-Text", "Mock STT capability inventory.", ContentStandard, SelectorPalette[3], "Devices|Pixel 2 AI Server", "State|mock available"),
		mockSection("Video Generation", "Mock video capability inventory.", ContentStandard, SelectorPalette[3], "Devices|Trash Can Mac Pro", "State|mock planned"),
	}
}

func researchMockSections() []Section {
	return []Section{
		selectorHeading("Your Research", "The total compute power you've donated to each type of research.", SelectorPalette[0]),
		mockSection("Curing Cancer (2.4 pflops)", "Mock contribution record; no work is running.", ContentStandard, SelectorPalette[0], "Contribution|2.4 pflops mock total", "Focus|medical research"),
		mockSection("Finding Aliens (6.2 gflops)", "Mock contribution record; no work is running.", ContentStandard, SelectorPalette[0], "Contribution|6.2 gflops mock total", "Focus|signal analysis"),
		selectorHeading("Other Research Opportunities", "Mock opportunities awaiting technical validation.", SelectorPalette[1]),
		mockSection("Drug Research", "Mock candidate research opportunity.", ContentStandard, SelectorPalette[1], "Status|not enrolled", "Budget|not assigned"),
		mockSection("Einstein@Home", "Analyze gravity waves to help find new neutron stars.", ContentStandard, SelectorPalette[1], "Status|not enrolled", "Evidence|mock review pending"),
		mockSection("NFS@Home", "Find new factorizations of large integers.", ContentStandard, SelectorPalette[1], "Status|not enrolled", "Evidence|mock review pending"),
	}
}

func mockSection(title, description string, kind ContentKind, color string, entries ...string) Section {
	return Section{Title: title, Description: description, Rows: mockRows(entries...), ContentKind: kind, SelectorColor: color}
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
