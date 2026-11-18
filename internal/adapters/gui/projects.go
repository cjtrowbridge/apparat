//go:build gui

// Presents registered checkouts and asynchronous folder and clone operations.
// All form state and selection belong to the GUI thread.
package gui

import (
	"context"
	"fmt"
	"image/color"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

type projectStore interface {
	Projects(context.Context) ([]domain.Project, error)
	StoreProject(context.Context, string, string, string) error
	CloneProject(context.Context, string, string, string) error
	RemoveProject(context.Context, string) error
	ProjectDefault(context.Context) (string, error)
	SetProjectDefault(context.Context, string) error
}
type projectResult struct {
	applyField                      func()
	workspace                       *workspaceReply
	rows                            []domain.Project
	folder, message, defaultAccount string
	reload                          bool
}
type projectForm struct {
	workspace                                workspacePanel
	dataRoot                                 string
	removeID                                 string
	workers                                  sync.WaitGroup
	rows                                     []domain.Project
	store                                    projectStore
	pending                                  chan projectResult
	cancel                                   context.CancelFunc
	ctx                                      context.Context
	close                                    context.CancelFunc
	repo, location, account, message, editID string
	active                                   bool
}

func (g *Game) initProjects() {
	f := &g.projects
	if runtime.GOOS == "android" && filepath.IsAbs(g.runtimeInfo.RuntimePath) {
		f.dataRoot = filepath.Join(g.runtimeInfo.RuntimePath, "data")
		// Registered filesystem checkouts already resolve symlinks at inspection.
		resolved, err := filepath.EvalSymlinks(f.dataRoot)
		if err == nil {
			f.dataRoot = resolved
		}
	}
	f.store, _ = g.runtimeInfo.Queries.(projectStore)
	f.ctx, f.close = context.WithCancel(context.Background())
}
func (g *Game) projectWork(action func(context.Context) projectResult) {
	f := &g.projects
	if f.pending != nil {
		return
	}
	ctx, cancel := context.WithCancel(f.ctx)
	f.cancel = cancel
	ch := make(chan projectResult, 1)
	f.pending = ch
	g.layoutDirty = true
	f.workers.Add(1)
	go func() {
		defer f.workers.Done()
		defer cancel()
		ch <- action(ctx)
	}()
}
func (g *Game) changeProjects(fn func(context.Context) error) {
	store := g.projects.store
	if store == nil {
		return
	}
	g.projectWork(func(ctx context.Context) projectResult {
		result := projectResult{}
		err := fn(ctx)
		if err != nil {
			result.message = err.Error()
		}
		rows, err := store.Projects(ctx)
		if err != nil {
			result.message = "Could not read project registrations"
			return result
		}
		result.rows = rows
		result.reload = true
		result.defaultAccount, _ = store.ProjectDefault(ctx)
		return result
	})
}
func (g *Game) pollProjects() {
	f := &g.projects
	active := g.shell.Snapshot().ActiveTab().ID() == hud.TabProjects
	if active && !f.active && f.pending == nil {
		g.changeProjects(func(context.Context) error {
			return nil
		})
	}
	f.active = active
	select {
	case r := <-f.pending:
		f.pending = nil
		f.message = r.message
		if r.applyField != nil {
			r.applyField()
		}
		if r.workspace != nil {
			g.applyWorkspace(r.workspace)
		}
		if r.reload {
			f.rows = r.rows
			if f.editID == "" {
				f.account = r.defaultAccount
			}
		}
		if r.folder != "" {
			f.location = r.folder
		}
		g.layoutDirty = true
	default:
	}
}
func (g *Game) projectTab(t hud.Tab) hud.Tab {
	t.Summary = ""
	t.Sections = []hud.Section{{Title: "Project repos", SelectorKind: hud.SelectorHeading, Description: "Working trees on this device."}, {Title: "Add or clone repo", SelectorColor: "#36ad9c", ContentKind: "project-form"}}
	for _, p := range g.projects.rows {
		kind := hud.ContentProject
		tint := hud.SelectorPalette[0]
		if len(p.Entrypoints) > 0 {
			kind = hud.ContentPipeline
		}
		configured := len(p.Actions)+len(p.Entrypoints) > 0
		if configured {
			tint = hud.SelectorPalette[1]
		}
		status := projectCheckoutStatus(p)
		section := hud.Section{Title: p.Name, SelectorSubtitle: projectListLocation(p, g.projects.dataRoot) + "\n" + status, Description: g.projectText(projectLocation(p.Location)), ContentKind: kind, ProjectID: p.ID, HasConfiguredActions: configured, SelectorColor: tint, Rows: []hud.Row{{Label: "Checkout", Detail: status}}}
		for _, task := range p.Entrypoints {
			section.Rows = append(section.Rows, hud.Row{Label: "Manual Task", Detail: task})
		}
		for _, action := range p.Actions {
			section.Rows = append(section.Rows, hud.Row{Label: "VS Code task label", Detail: action})
		}
		t.Sections = append(t.Sections, section)
	}
	return t
}
func (g *Game) projectButton(c *widget.Container, label string, click func()) {
	c.AddChild(widget.NewButton(widget.ButtonOpts.Text(label, g.theme.ButtonTheme.TextFace, g.theme.ButtonTheme.TextColor),
		widget.ButtonOpts.Image(g.theme.ButtonTheme.Image), widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(0, 44)),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			if !g.bodySelectionSuppressed() {
				click()
			}
		})))
}
func (g *Game) projectInput(c *widget.Container, label, initial string, changed func(string)) {
	c.AddChild(g.detailText(label))
	c.AddChild(g.newProjectInput(initial, changed))
}
func (g *Game) newProjectInput(initial string, changed func(string)) *widget.TextInput {
	w := widget.NewTextInput(widget.TextInputOpts.Face(g.theme.ButtonTheme.TextFace), widget.TextInputOpts.Image(&widget.TextInputImage{Idle: uiimage.NewNineSliceColor(backgroundColor), Highlight: uiimage.NewNineSliceColor(activeTabColor)}), widget.TextInputOpts.Color(&widget.TextInputColor{Idle: color.White, Caret: color.White}), widget.TextInputOpts.WidgetOpts(widget.WidgetOpts.MinSize(0, 44), widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})), widget.TextInputOpts.ChangedHandler(func(e *widget.TextInputChangedEventArgs) {
		changed(e.InputText)
	}))
	w.SetText(initial)
	return w
}
func (g *Game) projectControls(c *widget.Container, s hud.Section) {
	f := &g.projects
	store := f.store
	if store == nil {
		return
	}
	if f.message != "" {
		c.AddChild(g.detailText(g.projectText(f.message)))
	}
	if s.ProjectID != "" && f.workspace.id == s.ProjectID && f.workspace.mode != "registration" {
		g.workspaceControlsFor(c, s.ProjectID)
		return
	}
	if f.pending != nil {
		c.AddChild(g.detailText("Working…"))
		g.projectButton(c, "Cancel", f.cancel)
		return
	}
	if s.ProjectID != "" {
		if g.workspaceControlsFor(c, s.ProjectID) {
			return
		}
		for _, p := range f.rows {
			if p.ID == s.ProjectID {
				g.projectButton(c, "Edit registration", func() {
					f.editID = p.ID
					f.location = p.Location
					f.account = p.Account
					g.selectSection(hud.TabProjects, 1)
				})
				removeLabel := "Remove from list"
				if f.removeID == p.ID {
					removeLabel = "Confirm removal of project records"
				}
				g.projectButton(c, removeLabel, func() {
					if f.removeID != p.ID {
						f.removeID = p.ID
						g.layoutDirty = true
						return
					}
					f.removeID = ""
					g.changeProjects(func(ctx context.Context) error {
						return store.RemoveProject(ctx, p.ID)
					})
					delete(g.selectedSections, hud.TabProjects)
					g.detailOpen[hud.TabProjects] = false
				})
			}
		}
		return
	}
	g.projectField(c, "Folder", &f.location)
	if f.location != "" && f.editID == "" {
		g.projectButton(c, "Use default clone location", func() {
			f.location = ""
			f.message = ""
			g.layoutDirty = true
		})
	}
	g.projectButton(c, "Choose folder…", func() {
		g.projectWork(func(ctx context.Context) projectResult {
			picker := g.runtimeInfo.ProjectPicker
			if picker != nil {
				return projectResult{folder: picker.ChooseFolder()}
			}
			cmd := exec.CommandContext(ctx, "zenity", "--file-selection", "--directory")
			cmd.Env = append(os.Environ(), "GTK_USE_PORTAL=1")
			out, err := cmd.Output()
			if err != nil {
				return projectResult{message: "Folder selection closed or unavailable. You can enter a path."}
			}
			return projectResult{folder: strings.TrimSpace(string(out))}
		})
	})
	if runtime.GOOS != "android" {
		g.projectField(c, "GitHub login (blank for public)", &f.account)
		g.projectButton(c, "Use as device default", func() {
			account := f.account
			g.changeProjects(func(ctx context.Context) error {
				return store.SetProjectDefault(ctx, account)
			})
		})
	}
	save := "Open existing repo"
	if f.editID != "" {
		save = "Save registration"
	}
	g.projectButton(c, save, func() {
		id, path, account := f.editID, f.location, f.account
		g.changeProjects(func(ctx context.Context) error {
			return store.StoreProject(ctx, id, path, account)
		})
	})
	if f.editID != "" {
		g.projectButton(c, "New registration", func() {
			f.editID = ""
			f.location = ""
			f.message = ""
			g.layoutDirty = true
		})
		return
	}
	g.projectField(c, "GitHub owner/repo or URL", &f.repo)
	g.projectButton(c, "Preview clone", func() {
		repo, err := projectgit.RepoName(f.repo)
		if err != nil {
			f.message = err.Error()
			g.layoutDirty = true
			return
		}
		dest, err := g.projectDestination(f.location, repo)
		if err != nil {
			f.message = err.Error()
		} else {
			f.message = clonePreview(dest, f.account)
		}
		g.layoutDirty = true
	})
	g.projectButton(c, "Clone", func() {
		parent, repo, account := f.location, f.repo, f.account
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
func (g *Game) projectText(value string) string {
	return wrapSelectorLabel(value, *g.theme.ButtonTheme.TextFace, int(g.detailTextMaxWidth()))
}
func projectLocation(value string) string {
	if strings.HasPrefix(value, "content://") {
		_, id, ok := strings.Cut(value, "/document/")
		decoded, err := url.PathUnescape(id)
		if ok && err == nil {
			return decoded
		}
	}
	return value
}

// Registrations currently belong to this device. Provider document IDs can
// be opaque, so retain the provider and include the registered folder name.
func projectListLocation(p domain.Project, dataRoot string) string {
	location := p.Location
	if strings.HasPrefix(location, "content://") {
		id := projectLocation(location)
		if strings.HasPrefix(location, "content://com.android.externalstorage.documents/") {
			volume, path, ok := strings.Cut(id, ":")
			if ok {
				if volume == "primary" {
					volume = "~"
				}
				return "local://" + volume + "/" + path
			}
		}
		uri, err := url.Parse(location)
		if err == nil {
			location = uri.Host + "/" + id
			if !strings.HasSuffix(location, "/"+p.Name) {
				location += "/" + p.Name
			}
		}
	} else if dataRoot != "" && filepath.IsAbs(location) {
		rel, err := filepath.Rel(dataRoot, location)
		if err == nil && filepath.IsLocal(rel) {
			location = filepath.Join("apparat", rel)
		}
	}
	return "local://" + location
}

func projectCheckoutStatus(p domain.Project) string {
	if !p.Available {
		return "Unavailable · reconnect its folder"
	}
	branch := p.Branch
	if branch == "" {
		if p.Revision == "" {
			return "Checkout unavailable"
		}
		branch = "Detached HEAD"
	}
	revision := p.Revision
	if revision == "" {
		revision = "no resolved commit"
	}
	return branch + " @ " + revision
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
