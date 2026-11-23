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
	preview                                   domain.LocalProjectSetup
	progress                                  chan string
	output                                    []string
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
		if selected || strings.Contains(strings.ToLower(host.label), strings.ToLower(strings.TrimSpace(w.search))) {
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
		w.host = c.id
		w.remoteAccount = ""
		f.cloneParent, f.initParent, f.importPath = "", "", ""
		w.source = ""
		w.mode = ""
	case "source":
		if w.source == c.id {
			return
		}
		w.source = c.id
		w.mode = ""
		if c.id == "create" {
			w.mode = "starter"
		}
	case "mode":
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
	c.AddChild(g.detailText("Choose a device"))
	c.AddChild(g.detailText("Work takes place on the selected device. With remote projects, it will be able to use the cluster’s online resources, and cluster devices will be able to browse its contents while the host is online."))
	if g.runtimeInfo.PromptValue != nil {
		g.projectField(c, "Search devices", &w.search)
	} else {
		g.projectInput(c, "Search devices", w.search, func(value string) {
			if w.search != value {
				w.search = value
				g.layoutDirty = true
			}
		})
	}
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
			c.AddChild(g.detailText(label))
		}
		label := choice.label
		if choice.id == w.host || choice.id == w.source || choice.id == w.mode {
			label = "✓ " + label
		}
		g.projectButton(c, label, func() {
			g.chooseProject(choice)
		})
	}
	if len(choices) == 0 {
		c.AddChild(g.detailText("No matching devices."))
	}
	if step == "host" {
		return
	}
	local := w.host == "local"
	if !local {
		c.AddChild(g.detailText("Remote project setup is a preview. Execution on this host is not available yet."))
	}
	if w.mode == "starter" {
		c.AddChild(g.detailText("A blank project with agentic-pipelines and a repeatable bootstrap. Available once the public starter repository is published."))
		return
	}
	folder := func(label string, value *string, optional bool) {
		if local {
			g.projectFolder(c, label, value, optional)
		} else {
			g.projectField(c, label, value)
		}
	}
	switch w.mode {
	case "empty":
		g.projectField(c, "New repo name", &f.name)
		folder("Parent folder (optional)", &f.initParent, true)
	case "import":
		folder("Existing repo folder", &f.importPath, false)
	case "clone":
		g.projectField(c, "GitHub owner/repo or URL", &f.repo)
		folder("Destination parent folder (optional)", &f.cloneParent, true)
		if local && runtime.GOOS != "android" {
			g.projectAccount(c, &f.account)
		} else if local {
			c.AddChild(g.detailText("Public GitHub repositories only on Android."))
		} else {
			g.projectField(c, "GitHub login on the host (optional)", &w.remoteAccount)
		}
	default:
		return
	}
	if !local {
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
