//go:build gui

// Presents registered checkouts and asynchronous folder and clone operations.
// All form state and selection belong to the GUI thread.
package gui

import (
	"context"
	"image/color"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

type projectStore interface {
	Projects(context.Context) ([]domain.Project, error)
	StoreProject(context.Context, string, string, string) error
	SetupLocalProject(context.Context, domain.LocalProjectSetup, func(string)) (string, error)
	RemoveProject(context.Context, string) error
	ProjectDefault(context.Context) (string, error)
	SetProjectDefault(context.Context, string) error
}
type projectResult struct {
	applyField              func()
	workspace               *workspaceReply
	rows                    []domain.Project
	message, defaultAccount string
	reload                  bool
}
type projectForm struct {
	olderPages                                                 map[string]bool
	live                                                       liveDirectory
	wizard                                                     newProjectState
	directoryStatus                                            chan workspaceReply
	stopDirectoryStatus                                        context.CancelFunc
	externalCopies                                             map[[2]string]externalCopy
	copiedText                                                 string
	workspace                                                  workspacePanel
	dataRoot                                                   string
	removeID                                                   string
	workers                                                    sync.WaitGroup
	rows                                                       []domain.Project
	store                                                      projectStore
	pending                                                    chan projectResult
	cancel                                                     context.CancelFunc
	ctx                                                        context.Context
	close                                                      context.CancelFunc
	name, repo, account, defaultAccount, message, editID       string
	cloneParent, initParent, importPath, editPath, editAccount string
	active, catalogInitialized, catalogLoaded                  bool
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
	if (!f.catalogInitialized || active && !f.active) && f.pending == nil {
		f.catalogInitialized = true
		g.changeProjects(func(context.Context) error {
			return nil
		})
	}
	f.active = active
	g.pollWorkspaceLive()
	g.receiveProjectProgress()
	select {
	case status := <-f.wizard.jobUpdates:
		f.wizard.jobStatus = status
		g.layoutDirty = true
	default:
	}
	select {
	case r := <-f.pending:
		f.pending = nil
		f.wizard.jobUpdates = nil
		f.wizard.jobStatus = domain.GitStatus{}
		g.receiveProjectProgress()
		f.wizard.progress = nil
		f.message = r.message
		if r.applyField != nil {
			r.applyField()
		}
		if r.workspace != nil {
			g.applyWorkspace(r.workspace)
		}
		if r.reload {
			f.catalogLoaded = true
			f.rows = r.rows
			f.olderPages = nil

			f.defaultAccount = r.defaultAccount
		}
		g.layoutDirty = true
	default:
	}
}
func (g *Game) projectTab(t hud.Tab) hud.Tab {
	t.Summary = ""
	t.Sections = projectEntrySections[:]
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
		section := hud.Section{Title: g.projectTitle(p), SelectorSubtitle: projectListLocation(p, g.projects.dataRoot) + "\n" + status, Description: g.projectText(projectLocation(p.Location)), ContentKind: kind, ProjectID: p.ID, HasConfiguredActions: configured, SelectorColor: tint, Rows: []hud.Row{{Label: "Checkout", Detail: status}}}
		for _, task := range p.Entrypoints {
			section.Rows = append(section.Rows, hud.Row{Label: "Manual Task", Detail: task})
		}
		for _, action := range p.Actions {
			section.Rows = append(section.Rows, hud.Row{Label: "VS Code task label", Detail: action})
		}
		t.Sections = append(t.Sections, section)
	}
	for _, turn := range g.projects.pageTurns() {
		t.Sections = append(t.Sections, hud.Section{Title: turn.label, ActionID: "catalog-page/" + turn.host + "/" + turn.cursor, ContentKind: hud.ContentStandard})
	}
	return t
}
func (g *Game) projectButton(c *widget.Container, label string, click func()) *widget.Button {
	button := widget.NewButton(widget.ButtonOpts.Text(label, g.theme.ButtonTheme.TextFace, g.theme.ButtonTheme.TextColor),
		widget.ButtonOpts.Image(g.theme.ButtonTheme.Image), widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(0, 44)),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
			if !g.bodySelectionSuppressed() {
				click()
			}
		}))
	c.AddChild(button)
	return button
}
func (g *Game) projectInput(c *widget.Container, label, initial string, changed func(string)) *widget.TextInput {
	row := g.formFieldRow(c, label)
	field := g.editorForLayout(label, func() *widget.TextInput {
		return g.newProjectInput(initial, changed)
	})
	row.AddChild(field)
	return field
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

	if f.pending != nil && f.wizard.jobUpdates != nil {
		g.hostJobControls(c)
		return
	}
	if s.ProjectID != "" && f.workspace.id == s.ProjectID && f.workspace.mode != "registration" {
		g.workspaceControlsFor(c, s.ProjectID)
		return
	}
	if projectEntryAction(s.ContentKind) && len(f.wizard.output) > 0 {
		c.AddChild(g.detailText(strings.Join(f.wizard.output, "\n")))
		if f.wizard.progress != nil {
			g.newProjectControls(c)
			for _, child := range c.Children() {
				child.GetWidget().Disabled = true
			}
			return
		}
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
				if f.editID == p.ID {
					g.editProjectControls(c)
					return
				}
				g.projectButton(c, "Edit registration", func() {
					f.editID = p.ID
					f.editPath = p.Location
					f.editAccount = p.Account
					g.layoutDirty = true
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
	if projectEntryAction(s.ContentKind) {
		g.newProjectControls(c)
	}
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
	if p.RemoteHost != "" {
		return p.RemoteHost + "://" + p.Location
	}
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
		return branch + " · No commits yet"
	}
	return branch + " @ " + revision
}

func (g *Game) selectProjectEntry(index int) bool {
	f := &g.projects
	if f.pending != nil {
		return false
	}
	f.wizard.focusStep = ""
	if index == 1 {
		f.wizard.focusStep = "search"
	}
	if index != g.selectedSections[hud.TabProjects] {
		f.editID = ""
		f.message = ""
	}
	row := index - len(projectEntrySections)
	if f.store != nil && row >= len(f.rows) {
		turns := f.pageTurns()
		i := row - len(f.rows)
		if i < len(turns) {
			g.turnProjectPage(turns[i])
		}
		return false
	}
	if row >= 0 && row < len(f.rows) {
		p := f.rows[row]
		if p.Available && f.workspace.id != p.ID {
			f.workspace = workspacePanel{id: p.ID, dir: ".", mode: "files"}
			g.detailOpen[hud.TabProjects] = true
			g.workspaceRequest("list")
		}
	}
	return true
}
