//go:build gui

// Creates an offline project and selects its registered workspace on success.
// GitHub clone previews share the same destination controls.
package gui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
)

func clonePreview(dest, account string) string {
	if account == "" {
		account = "public access"
	}
	return fmt.Sprintf("Clone to %s (%s)", dest, account)
}

func (g *Game) projectDestination(parent, repo string) (string, error) {
	if strings.HasPrefix(parent, "content://") {
		return projectLocation(parent) + "/" + repo[strings.IndexByte(repo, '/')+1:], nil
	}
	return projectgit.Destination(g.runtimeInfo.RuntimePath, parent, repo)
}

const newRepoForm hud.ContentKind = "new-project"

var projectEntrySections = [...]hud.Section{
	{Title: "Project repos", SelectorKind: hud.SelectorHeading, Description: "Working trees on this device."},
	{Title: "New Project", ContentKind: newRepoForm, SelectorColor: "#36ad9c"},
}

func (g *Game) projectFolder(c *widget.Container, label string, value *string, optional bool) widget.Focuser {
	field := g.projectField(c, label, value)
	if optional && *value != "" {
		g.projectButton(c, "Use default repo location", func() {
			*value = ""
			g.projects.message = ""
			g.layoutDirty = true
		})
	}
	g.projectButton(c, "Choose folder…", func() {
		g.projectWork(func(ctx context.Context) projectResult {
			var path string
			picker := g.runtimeInfo.ProjectPicker
			if picker != nil {
				path = picker.ChooseFolder()
			} else {
				cmd := exec.CommandContext(ctx, "zenity", "--file-selection", "--directory")
				cmd.Env = append(os.Environ(), "GTK_USE_PORTAL=1")
				out, err := cmd.Output()
				if err != nil {
					return projectResult{message: "Folder selection closed or unavailable. You can enter a path."}
				}
				path = strings.TrimSpace(string(out))
			}
			return projectResult{applyField: func() {
				if path != "" {
					*value = path
				}
			}}
		})
	})
	return field
}

func (g *Game) projectAccount(c *widget.Container, value *string) {
	if runtime.GOOS != "android" {
		g.hostAccountPicker(c, "local", value)
		g.projectButton(c, "Use as device default", func() {
			account := *value
			g.changeProjects(func(ctx context.Context) error {
				return g.projects.store.SetProjectDefault(ctx, account)
			})
		})
	}
}

func (g *Game) editProjectControls(c *widget.Container) {
	f := &g.projects
	g.projectFolder(c, "Registered repo folder", &f.editPath, false)
	g.projectAccount(c, &f.editAccount)
	g.projectButton(c, "Save registration", func() {
		id, path, account := f.editID, f.editPath, f.editAccount
		g.changeProjects(func(ctx context.Context) error {
			return f.store.StoreProject(ctx, id, path, account)
		})
	})
	g.projectButton(c, "Done editing", func() {
		f.editID = ""
		f.message = ""
		g.layoutDirty = true
	})
}

func (g *Game) hostAccountPicker(c *widget.Container, host string, value *string) {
	c.AddChild(g.formLabel("GitHub account"))
	for _, option := range g.accountOptions(host, value) {
		g.projectButton(c, option.label, option.run)
	}
}

type accountOption struct {
	id, label string
	run       func()
}

func (g *Game) accountOptions(host string, value *string) []accountOption {
	var options []accountOption
	add := func(id, label string, run func()) {
		options = append(options, accountOption{id, label, run})
	}
	w := &g.projects.wizard
	current := *value
	if current == "" {
		current = "Public access"
	}
	add("choose", current+" · Choose account", func() {
		source, ok := g.runtimeInfo.Queries.(interface {
			GitAccounts(context.Context, string) ([]string, error)
		})
		if !ok {
			return
		}
		w.accounts = nil
		w.accountHost = ""
		g.projectWork(func(ctx context.Context) projectResult {
			names, err := source.GitAccounts(ctx, host)
			if err != nil {
				return projectResult{message: "Accounts unavailable. Verify host trust and administrator access."}
			}
			return projectResult{applyField: func() {
				w.accounts = names
				w.accountHost = host
			}}
		})
	})
	if w.accountHost != host {
		return options
	}
	selectAccount := func(name string) {
		*value = name
		w.accountHost = ""
		w.accounts = nil
		w.preview = domain.LocalProjectSetup{}
		g.layoutDirty = true
	}
	add("public", "Public access", func() {
		selectAccount("")
	})
	for _, name := range w.accounts {
		add("user-"+name, name, func() {
			selectAccount(name)
		})
	}
	add("add", "➕ Add account on host", func() {
		if host == "local" {
			g.projectWork(func(ctx context.Context) projectResult {
				err := projectgit.AddLocalAccount(ctx)
				if err != nil {
					return projectResult{message: "Open a host terminal, run gh auth login --hostname github.com, then refresh the account list."}
				}
				return projectResult{message: "Choose account again to refresh the list."}
			})
			return
		}
		g.projects.message = "Run gh auth login --hostname github.com in a terminal as the host account running Apparat. Then choose an account again."
		g.layoutDirty = true
	})
	return options
}
