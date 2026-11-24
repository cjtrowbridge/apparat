//go:build gui

// Keeps viewing, editing and draft recovery in one Projects detail pane.
// Every write uses the owner and base revision captured with the document.
package gui

import (
	"github.com/ebitenui/ebitenui/widget"
	"strings"
)

func (g *Game) fileDocument(c *widget.Container, w *workspacePanel) {
	c.AddChild(g.detailText(w.path))
	if w.surface == nil {
		w.surface = g.newDocumentSurface(w)
	} else {
		reattachEditor(w.surface.TextInput)
	}
	if !w.editing {
		if w.readOnly {
			c.AddChild(g.detailText("Read-only · Cluster administrators can edit this file."))
		} else {
			g.projectButton(c, "Edit", func() {
				g.beginDocumentEdit(w)
			})
			g.externalDocumentControls(c, w)
		}
	} else {
		label := "Save"
		for _, p := range g.projects.rows {
			if p.ID == w.id && p.RemoteHost != "" {
				label = "Save to " + p.RemoteHost
				break
			}
		}
		g.projectButton(c, label, func() {
			if g.projects.pending == nil {
				g.workspaceRequest("save")
			}
		})
		cancel := "Cancel"
		if w.confirmation == "cancel-edit" {
			cancel = "Discard changes?"
		}
		g.projectButton(c, cancel, func() {
			g.cancelDocumentEdit(w)
		})
		if w.report != "" && w.compared == "" {
			g.projectButton(c, "Compare with saved file", func() {
				g.workspaceRequest("compare")
			})
		}
		if w.compared != "" {
			g.projectButton(c, "Keep draft against reviewed revision", func() {
				g.workspaceRequest("rebase")
			})
		}
	}
	if w.editing {
		g.multilineRow(c, "Document").AddChild(w.surface)
	} else {
		c.AddChild(w.surface)
	}
	tools := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(2))))
	c.AddChild(tools)
	g.projectButton(tools, "Select all", func() {
		w.surface.anchor = 0
		w.surface.caret = len(w.surface.body)
		w.surface.syncIME()
	})
	g.projectButton(tools, "Copy selection", func() {
		s := w.surface
		g.copyDocument(s.body[min(s.anchor, s.caret):max(s.anchor, s.caret)])
	})
	if w.editing {
		g.projectButton(tools, "Undo", func() {
			w.surface.history(&w.surface.undo, &w.surface.redo)
		})
		g.projectButton(tools, "Redo", func() {
			w.surface.history(&w.surface.redo, &w.surface.undo)
		})
		g.projectButton(tools, "Paste", func() {
			w.surface.insert(g.pasteDocument())
		})
	}
}

func (g *Game) beginDocumentEdit(w *workspacePanel) {
	if w.readOnly || w.editing || g.projects.pending != nil {
		return
	}
	w.originalBase = w.base
	w.original = strings.Join(w.lines, "\n")
	w.editing = true
	w.confirmation = ""
	w.report = ""
	if w.surface != nil {
		w.surface.undo = nil
		w.surface.redo = nil
	}
	g.layoutDirty = true
}
func (g *Game) cancelDocumentEdit(w *workspacePanel) {
	if g.projects.pending != nil {
		return
	}
	if strings.Join(w.lines, "\n") != w.original {
		if w.confirmation != "cancel-edit" {
			w.confirmation = "cancel-edit"
			g.layoutDirty = true
			return
		}
		oldBase, oldLines := w.base, w.lines
		w.base = w.originalBase
		w.lines = strings.Split(w.original, "\n")
		if !g.keepEditingDraft(w) {
			w.base, w.lines = oldBase, oldLines
			return
		}
	}
	w.editing = false
	if w.surface != nil {
		w.surface.ime.Cancel()
	}
	w.surface = nil
	w.confirmation = ""
	w.report = ""
	w.compared = ""
	g.layoutDirty = true
}
