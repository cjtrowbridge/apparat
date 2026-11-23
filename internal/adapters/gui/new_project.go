//go:build gui

// Presents a progressive repo setup form using the synchronized device directory.
// Only the explicit local destination can invoke the checkout store.
package gui

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
	"runtime"
	"strings"
)

type newProjectState struct {
	search, host, source, mode, remoteAccount string
	accountHost                               string
	providers                                 []domain.GitProvider
	providerError                             string
	jobUpdates                                chan domain.GitStatus
	jobStatus                                 domain.GitStatus
	retryRequest                              domain.GitRequest
	preview                                   domain.LocalProjectSetup
	progress                                  chan string
	output                                    []string
	focusStep                                 string
	focusControl                              widget.Focuser
}
type setupChoice struct{ id, label, step string }

func (g *Game) projectHosts() []setupChoice {
	local := "This device"
	hosts := []setupChoice{{"local", local, "host"}}
	for _, s := range g.personalSections() {
		if s.CurrentDevice {
			hosts[0].label = s.Title + " (this device)"
		} else {
			hosts = append(hosts, setupChoice{s.ActionID, s.Title, "host"})
		}
	}
	for _, w := range g.serviceProjection.Devices {
		hosts = append(hosts, setupChoice{"worker-" + w.Fingerprint, w.DisplayName(), "host"})
	}
	return hosts
}
func (g *Game) projectChoices() []setupChoice {
	w := &g.projects.wizard
	var choices []setupChoice
	found := false
	for _, host := range g.projectHosts() {
		selected := w.host == host.id
		found = found || selected
		if selected || matchesProjectHost(host, w.search) {
			choices = append(choices, host)
		}
	}
	if !found {
		return choices
	}
	choices = append(choices, setupChoice{"create", "Create a new repository", "source"}, setupChoice{"bring", "Bring in a repository", "source"})
	switch w.source {
	case "create":
		choices = append(choices, setupChoice{"starter", "Agentic Pipelines starter (default)", "mode"}, setupChoice{"empty", "Empty Git repository", "mode"})
	case "bring":
		choices = append(choices, setupChoice{"clone", "Clone from GitHub", "mode"}, setupChoice{"import", "Existing folder", "mode"})
	}
	return choices
}
func (g *Game) chooseProject(c setupChoice) {
	f := &g.projects
	if f.pending != nil {
		return
	}
	w := &f.wizard
	switch c.step {
	case "host":
		if w.host == c.id {
			return
		}
		w.focusStep = "source"
		w.host = c.id
		w.remoteAccount = ""
		w.providers = nil
		w.retryRequest = domain.GitRequest{}
		w.providerError = ""
		w.accountHost = ""
		f.account = ""
		f.cloneParent, f.initParent, f.importPath = "", "", ""
		w.source = ""
		w.mode = ""
	case "source":
		if w.source == c.id {
			return
		}
		w.focusStep = "mode"
		w.source = c.id
		w.mode = ""
		if c.id == "create" {
			w.mode = "starter"
		}
	case "mode":
		w.focusStep = "details"
		w.mode = c.id
	}
	w.preview = domain.LocalProjectSetup{}
	w.output = nil
	f.message = ""
	g.layoutDirty = true
}

func (g *Game) newProjectControls(c *widget.Container) {
	f := &g.projects
	w := &f.wizard
	c.AddChild(g.formLabel("Choose a device"))
	c.AddChild(g.formDescription("Work takes place on the selected device. With remote projects, it will be able to use the cluster’s online resources, and cluster devices will be able to browse its contents while the host is online."))
	g.deviceSearchField(c)
	choices := g.projectChoices()
	step := "host"
	for _, choice := range choices {
		if choice.step != step {
			step = choice.step
			label := "Repository source"
			if step == "mode" {
				label = "Import method"
				if w.source == "create" {
					label = "Project template"
				}
			}
			c.AddChild(g.formLabel(label))
		}
		label := choice.label
		if choice.id == w.host || choice.id == w.source || choice.id == w.mode {
			label = "✓ " + label
		}
		button := g.projectButton(c, label, func() {
			g.chooseProject(choice)
		})
		g.projectFocusControl(choice.step, button)
	}
	if len(choices) == 0 {
		c.AddChild(g.formDescription("No matching devices."))
	}
	if step == "host" {
		return
	}
	local := w.host == "local"
	if !local && (!strings.HasPrefix(w.host, "worker-") || w.mode == "import" || w.mode == "empty") {
		c.AddChild(g.formDescription("Remote project setup is a preview. Execution on this host is not available yet."))
	}
	if w.mode == "starter" {
		c.AddChild(g.formDescription("A blank project with agentic-pipelines and a repeatable bootstrap. Available once the public starter repository is published."))
		return
	}
	folder := func(label string, value *string, optional bool) widget.Focuser {
		if local {
			return g.projectFolder(c, label, value, optional)
		}
		return g.projectField(c, label, value)
	}
	switch w.mode {
	case "empty":
		g.projectFocusControl("details", g.projectField(c, "New repo name", &f.name))
		folder("Parent folder (optional)", &f.initParent, true)
	case "import":
		g.projectFocusControl("details", folder("Existing repo folder", &f.importPath, false))
	case "clone":
		g.projectFocusControl("details", g.projectField(c, "GitHub owner/repo or URL", &f.repo))
		if local {
			folder("Destination parent folder (optional)", &f.cloneParent, true)
		} else if strings.HasPrefix(w.host, "worker-") {
			c.AddChild(g.formDescription("The worker saves the checkout in apparat/data/repos/<owner>/<repository>."))
		}
		if local && runtime.GOOS != "android" {
			g.projectAccount(c, &f.account)
		} else if local {
			g.hostAccountPicker(c, "local", &f.account)
		} else {
			if strings.HasPrefix(w.host, "worker-") {
				g.hostAccountPicker(c, w.host, &w.remoteAccount)
			} else {
				c.AddChild(g.formDescription("Account selection requires a hosting worker."))
			}
		}
	default:
		return
	}
	if !local {
		action := g.remoteCloneOption()
		if action != nil {
			g.projectButton(c, action.label, action.run)
		}
		return
	}
	if w.mode == "clone" {
		g.projectButton(c, "Preview clone", func() {
			input := g.projectSetupRequest()
			repo, err := projectgit.RepoName(input.Name)
			if err == nil {
				var dest string
				dest, err = g.projectDestination(input.Parent, repo)
				if err == nil {
					w.preview = input
					f.message = clonePreview(dest, input.Account)
				}
			}
			if err != nil {
				f.message = err.Error()
				w.preview = domain.LocalProjectSetup{}
			}
			g.layoutDirty = true
		})
	}
	req := g.projectSetupRequest()
	if req.Mode == "clone" && req != w.preview {
		return
	}
	g.projectButton(c, map[string]string{"clone": "Clone", "import": "Import repo", "empty": "Initialize repo"}[w.mode], g.startLocalProject)
}
func (g *Game) projectSetupRequest() domain.LocalProjectSetup {
	f := &g.projects
	req := domain.LocalProjectSetup{Mode: f.wizard.mode}
	switch req.Mode {
	case "empty":
		req.Name = f.name
		req.Parent = f.initParent
	case "import":
		req.Name = f.importPath
		req.Account = f.defaultAccount
	case "clone":
		req.Name = f.repo
		req.Parent = f.cloneParent
		req.Account = f.account
	}
	return req
}
func (g *Game) receiveProjectProgress() {
	w := &g.projects.wizard
	for {
		select {
		case msg := <-w.progress:
			w.output = append(w.output, msg)
			g.layoutDirty = true
		default:
			return
		}
	}
}
func (g *Game) startLocalProject() {
	f := &g.projects
	w := &f.wizard
	req := g.projectSetupRequest()
	if f.pending != nil || w.host != "local" || (req.Mode != "empty" && req.Mode != "import" && req.Mode != "clone") {
		return
	}
	if strings.TrimSpace(req.Name) == "" || req.Mode == "clone" && req != w.preview {
		f.message = "Complete the project details before starting."
		g.layoutDirty = true
		return
	}
	progress := make(chan string, 8)
	w.progress = progress
	w.output = []string{"Starting local project setup…"}
	f.message = ""
	store := f.store
	g.projectWork(func(ctx context.Context) projectResult {
		path, err := store.SetupLocalProject(ctx, req, func(msg string) {
			select {
			case progress <- msg:
			case <-ctx.Done():
			}
		})
		if err != nil {
			return projectResult{message: err.Error()}
		}
		progress <- "Refreshing registered projects and opening the checkout…"
		rows, err := store.Projects(ctx)
		if err != nil {
			return projectResult{message: "Registration saved. Reopen Projects to refresh the list."}
		}
		for i, p := range rows {
			if p.Location == path {
				return projectResult{applyField: func() {
					f.rows = rows
					f.catalogLoaded = true
					f.workspace = workspacePanel{}
					if g.shell.Snapshot().ActiveTab().ID() != hud.TabProjects {
						for j, tab := range g.shell.Snapshot().Tabs {
							if tab.ID() == hud.TabProjects {
								g.selectTab(j, nil)
								break
							}
						}
					}
					f.wizard = newProjectState{}
					f.name, f.repo, f.importPath = "", "", ""
					g.selectSection(hud.TabProjects, len(projectEntrySections)+i)
				}}
			}
		}
		return projectResult{message: "Registered checkout could not be located in the refreshed list."}
	})
}

var searchBreaks = strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ", "\u0085", " ", "\u2028", " ", "\u2029", " ")

func (g *Game) deviceSearchField(c *widget.Container) {
	const label = "Search devices"
	w := &g.projects.wizard
	if g.runtimeInfo.PromptValue != nil {
		button := g.nativeProjectField(c, label, w.search, true, g.filterProjectHosts)
		g.projectFocusControl("search", button)
		return
	}
	row := g.formFieldRow(c, label)
	field := g.editorForLayout(label, func() *widget.TextInput {
		input := g.newProjectInput(w.search, g.filterProjectHosts)
		widget.TextInputOpts.Validation(func(value string) (bool, *string) {
			line := searchLine(value)
			if line != value {
				return false, &line
			}
			return true, nil
		})(input)
		compactSearchControl(input.GetWidget(), g.prefs.Font)
		return input
	})
	row.AddChild(field)
	g.projectFocusControl("search", field)
}

func (g *Game) filterProjectHosts(value string) {
	w := &g.projects.wizard
	value = searchLine(value)
	if value == w.search {
		return
	}
	w.search = value
	g.layoutDirty = true
	if strings.TrimSpace(value) == "" || g.projects.pending != nil {
		return
	}
	var match setupChoice
	for _, host := range g.projectHosts() {
		if !matchesProjectHost(host, value) {
			continue
		}
		if match.id != "" {
			return
		}
		match = host
	}
	if match.id != "" {
		g.chooseProject(match)
		w.focusStep = "source"
	}
}

func (g *Game) projectFocusControl(step string, control widget.Focuser) {
	w := &g.projects.wizard
	if w.focusStep == step && w.focusControl == nil {
		w.focusControl = control
	}
}

// Apply after layout so a focus change can reveal the next control in its viewport.
func (g *Game) applyProjectFocus() {
	w := &g.projects.wizard
	if w.focusControl == nil || g.nav.owner != hud.TabProjects || g.actions.mode != "" {
		return
	}
	targets := g.bodyTargets()
	for i, target := range targets {
		if target.w != w.focusControl {
			continue
		}
		g.focusBody(targets, i)
		field, editing := target.w.(*widget.TextInput)
		if editing && w.focusStep == "search" {
			g.ui.SetFocusedWidget(field)
			field.SelectAll()
		}
		w.focusStep = ""
		w.focusControl = nil
		return
	}
}
