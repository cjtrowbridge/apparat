//go:build gui

// Projects the filesystem and Git index into one directory of navigable names.
// Missing tracked paths remain available for deletion review.
package gui

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/domain"
	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"image/color"
	"path"
	"sort"
	"strings"
)

type directoryItem struct {
	domain.WorkspaceEntry
	change  domain.WorkspaceChange
	missing bool
}

func directoryItems(dir string, entries []domain.WorkspaceEntry, changes []domain.WorkspaceChange) []directoryItem {
	children := map[string]directoryItem{}
	for _, e := range entries {
		children[e.Name] = directoryItem{WorkspaceEntry: e}
	}
	prefix := ""
	if dir != "." {
		prefix = dir + "/"
	}
	for _, ch := range changes {
		relative, ok := strings.CutPrefix(ch.Path, prefix)
		if !ok {
			continue
		}
		name, _, nested := strings.Cut(relative, "/")
		item, exists := children[name]
		if !exists {
			if ch.Marker != "D" {
				continue
			}
			item = directoryItem{WorkspaceEntry: domain.WorkspaceEntry{Name: name, Directory: nested || ch.Directory}, missing: true}
		}
		if !nested && !ch.Directory {
			item.change = ch
		}
		children[name] = item
	}
	rows := make([]directoryItem, 0, len(children))
	for _, item := range children {
		rows = append(rows, item)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Directory != rows[j].Directory {
			return rows[i].Directory
		}
		return rows[i].Name < rows[j].Name
	})
	return rows
}
func (g *Game) directoryListing(c *widget.Container, w *workspacePanel) {
	list := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))), widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})))
	c.AddChild(list)
	location := "/"
	if w.dir != "." {
		location += w.dir + "/"
	}
	if g.projects.directoryStatus != nil {
		location += "   Checking Git status…"
	}
	list.AddChild(g.detailText(location))
	if w.dir != "." {
		g.filenameRow(list, "../", false, g.workspaceBack)
	}
	rows := directoryItems(w.dir, w.entries, w.changes)
	for _, entry := range rows[:min(w.page+24, len(rows))] {
		label := entry.Name
		if entry.Directory {
			label += "/"
		} else if entry.change.Marker != "" {
			label = entry.change.Marker + "  " + label
		}
		g.filenameRow(list, label, entry.change.Marker != "", func() {
			dest := path.Join(w.dir, entry.Name)
			switch {
			case entry.Directory:
				if w.virtualDirs == nil {
					w.virtualDirs = map[string]bool{}
				}
				w.virtualDirs[dest] = entry.missing
				w.dir = dest
				w.page = 0
				g.workspaceGo("files", "list")
			case entry.missing:
				w.path = dest
				w.change = entry.change
				w.base = ""
				g.workspaceGo("review", "review")
			default:
				w.path = dest
				w.base = ""
				w.line = 0
				w.autoEdit = true
				g.workspaceRequest("read")
			}
		})
	}
	if w.cursor != "" && w.page+24 >= len(rows) {
		g.projectButton(c, "Next directory page", func() {
			g.workspaceRequest("next-directory")
		})
	}
	if w.page+24 < len(rows) {
		g.projectButton(c, "More files", func() {
			w.page += 24
			g.layoutDirty = true
		})
	}
}
func (g *Game) filenameRow(c *widget.Container, label string, changed bool, click func()) {
	// Keep keyboard/controller focus and a full-width touch target on the plain list row.
	tint := color.RGBA{255, 255, 255, 255}
	if changed {
		tint = color.RGBA{255, 145, 145, 255}
	}
	colors := &widget.ButtonTextColor{Idle: tint, Hover: tint, Pressed: tint, Disabled: tint}
	bg := *g.theme.ButtonTheme.Image
	bg.Idle = uiimage.NewNineSliceColor(color.Transparent)
	label = wrapSelectorLabel(label, *g.theme.ButtonTheme.TextFace, max(80, g.detailPanePreferredWidth()-52))
	c.AddChild(widget.NewButton(widget.ButtonOpts.Text(label, g.theme.ButtonTheme.TextFace, colors), widget.ButtonOpts.Image(&bg), widget.ButtonOpts.TextPadding(&widget.Insets{Left: 4, Right: 4}), widget.ButtonOpts.TextPosition(widget.TextPositionStart, widget.TextPositionCenter), widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(0, 44), widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})), widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
		if g.projects.pending == nil && !g.bodySelectionSuppressed() {
			click()
		}
	})))
}

func (g *Game) stopFolderCheck() {
	f := &g.projects
	if f.stopDirectoryStatus != nil {
		f.stopDirectoryStatus()
		f.stopDirectoryStatus = nil
	}
	f.directoryStatus = nil
}
func (g *Game) startDirectoryStatus(q domain.WorkspaceRequest) {
	g.stopFolderCheck()
	store, ok := g.runtimeInfo.Queries.(workspaceStore)
	if !ok {
		return
	}
	f := &g.projects
	ctx, cancel := context.WithCancel(f.ctx)
	ch := make(chan workspaceReply, 1)
	f.directoryStatus = ch
	f.stopDirectoryStatus = cancel
	q.Action = "directory-status"
	f.workers.Add(1)
	go func() {
		defer f.workers.Done()
		defer cancel()
		v, err := store.Workspace(ctx, q)
		ch <- workspaceReply{q, v, err}
	}()
}
func (g *Game) pollDirectoryStatus() {
	f := &g.projects
	select {
	case r := <-f.directoryStatus:
		g.stopFolderCheck()
		w := &f.workspace
		if w.mode == "files" && w.id == r.request.ProjectID && w.dir == r.request.Path {
			w.changes = r.view.Changes
			if r.err != nil {
				w.report = "Git status unavailable: " + r.err.Error()
			}
			g.layoutDirty = true
		}
	default:
	}
}
