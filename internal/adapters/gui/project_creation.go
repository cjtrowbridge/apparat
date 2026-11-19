//go:build gui

// Creates an offline project and selects its registered workspace on success.
// GitHub clone previews share the same destination controls.
package gui

import (
	"context"
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
	"strings"
)

func (g *Game) createProjectControls(c *widget.Container) {
	f := &g.projects
	g.projectField(c, "New project name", &f.name)
	g.projectButton(c, "Create project", func() {
		name, parent := f.name, f.location
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
						g.selectSection(hud.TabProjects, i+2)
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
