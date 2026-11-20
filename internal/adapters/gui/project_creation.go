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
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
)

func (g *Game) createProjectControls(c *widget.Container) {
	f := &g.projects
	g.projectField(c, "New repo name", &f.name)
	g.projectFolder(c, "Parent folder (optional)", &f.initParent, true)
	g.projectButton(c, "Initialize repo", func() {
		name, parent := f.name, f.initParent
		store := f.store
		g.projectWork(func(ctx context.Context) projectResult {
			path, err := store.CreateProject(ctx, parent, name)
			if err != nil {
				return projectResult{message: err.Error()}
			}
			rows, err := store.Projects(ctx)
			if err != nil {
				return projectResult{message: "Project created. Reopen Projects to refresh the list."}
			}
			return projectResult{applyField: func() {
				f.rows = rows
				f.name = ""
				for i, p := range rows {
					if p.Location == path {
						g.selectSection(hud.TabProjects, len(projectEntrySections)+i)
						break
					}
				}
			}}
		})
	})
}
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

const (
	cloneRepoForm  hud.ContentKind = "clone-repo"
	initRepoForm   hud.ContentKind = "init-repo"
	importRepoForm hud.ContentKind = "import-repo"
)

var projectEntrySections = [...]hud.Section{
	{Title: "Project repos", SelectorKind: hud.SelectorHeading, Description: "Working trees on this device."},
	{Title: "Clone remote repo", Description: "Download a GitHub repository to this device.", SelectorColor: "#36ad9c", ContentKind: cloneRepoForm},
	{Title: "Init new repo", Description: "Create an empty Git repository in a new folder.", SelectorColor: "#36ad9c", ContentKind: initRepoForm},
	{Title: "Import local repo", Description: "Add an existing Git working tree to Projects in its current location.", SelectorColor: "#36ad9c", ContentKind: importRepoForm},
}

func (g *Game) projectFolder(c *widget.Container, label string, value *string, optional bool) {
	g.projectField(c, label, value)
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
}

func (g *Game) projectAccount(c *widget.Container, value *string) {
	if runtime.GOOS != "android" {
		g.projectField(c, "GitHub login (blank for public)", value)
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

func (g *Game) cloneProjectControls(c *widget.Container) {
	f := &g.projects
	store := f.store
	g.projectButton(c, "Preview clone", func() {
		repo, err := projectgit.RepoName(f.repo)
		if err != nil {
			f.message = err.Error()
			g.layoutDirty = true
			return
		}
		dest, err := g.projectDestination(f.cloneParent, repo)
		if err != nil {
			f.message = err.Error()
		} else {
			f.message = clonePreview(dest, f.account)
		}
		g.layoutDirty = true
	})
	g.projectButton(c, "Clone", func() {
		parent, repo, account := f.cloneParent, f.repo, f.account
		normalized, err := projectgit.RepoName(repo)
		if err != nil {
			f.message = err.Error()
			g.layoutDirty = true
			return
		}
		dest, err := g.projectDestination(parent, normalized)
		if err != nil || f.message != clonePreview(dest, account) {
			f.message = "Preview the destination and account first."
			g.layoutDirty = true
			return
		}
		g.changeProjects(func(ctx context.Context) error {
			return store.CloneProject(ctx, parent, normalized, account)
		})
	})
}
