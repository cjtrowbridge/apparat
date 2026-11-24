//go:build gui

// Exercises view-first document transitions and scoped saves from the canvas.
// Includes external working-copy import, permission gating, cancellation and undo.
package gui

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

type documentFixture struct {
	ServiceQueries
	root string
	last domain.WorkspaceRequest
}

func (f *documentFixture) Workspace(ctx context.Context, q domain.WorkspaceRequest) (domain.WorkspaceView, error) {
	if q.Action == "save" {
		f.last = q
	}
	if q.Action == "draft" {
		return domain.WorkspaceView{}, nil
	}
	return projectgit.Files(f.root, q)
}
func TestDocumentViewEditSave(t *testing.T) {
	g := NewGame()
	defer g.Close()
	root := t.TempDir()
	q := domain.WorkspaceRequest{Action: "read", ProjectID: "worker-fixture/project", Path: "folder/note"}
	err := os.Mkdir(filepath.Join(root, "folder"), 0700)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "folder", "note")
	err = os.WriteFile(path, []byte("original"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	store := &documentFixture{root: root}
	g.runtimeInfo.Queries = store
	w := &g.projects.workspace
	w.id = q.ProjectID
	v, err := store.Workspace(t.Context(), q)
	if err != nil {
		t.Fatal(err)
	}
	g.applyWorkspace(&workspaceReply{request: q, view: v})
	w.surface = g.newDocumentSurface(w)
	e := w.surface
	e.insert("oops")
	if w.mode != "file" || w.editing || e.body != "original" || g.projects.pending != nil {
		t.Fatal("opening file must only view")
	}
	g.beginDocumentEdit(w)
	e.caret = len(e.body)
	e.anchor = e.caret
	e.insert("\nsecond")
	e.history(&e.undo, &e.redo)
	if e.body != "original" {
		t.Fatal("undo")
	}
	e.history(&e.redo, &e.undo)
	g.cancelDocumentEdit(w)
	if w.confirmation != "cancel-edit" || !w.editing {
		t.Fatal("dirty cancel skipped confirmation")
	}
	g.cancelDocumentEdit(w)
	if w.editing || w.lines[0] != "original" {
		t.Fatal("cancel failed")
	}
	w.surface = g.newDocumentSurface(w)
	g.beginDocumentEdit(w)
	w.surface.caret = len(w.surface.body)
	w.surface.insert("one\ntwo")
	g.workspaceRequest("save")
	result := <-g.projects.pending
	g.projects.pending = nil
	if result.workspace.err != nil {
		t.Fatal(result.workspace.err)
	}
	g.applyWorkspace(result.workspace)
	if store.last.ProjectID != q.ProjectID || store.last.Path != q.Path || store.last.Base != v.Base || w.mode != "files" || w.dir != "folder" {
		t.Fatal("save lost checkout identity, opening revision or directory")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "one\ntwo" {
		t.Fatal("incorrect saved text", err)
	}
	// A returned copy must retain the opening revision even if another writer saved meanwhile.
	listing := <-g.projects.pending
	g.projects.pending = nil
	g.applyWorkspace(listing.workspace)
	v, err = store.Workspace(t.Context(), q)
	if err != nil {
		t.Fatal(err)
	}
	g.applyWorkspace(&workspaceReply{request: q, view: v})
	copy, err := newExternalCopy("returned", v.Base)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(copy.dir)
	err = os.WriteFile(path, []byte("concurrent"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	g.importDocument(w, [2]string{w.id, w.path}, copy)
	g.workspaceRequest("save")
	result = <-g.projects.pending
	g.projects.pending = nil
	if result.workspace.err == nil {
		t.Fatal("stale external copy overwrote the checkout")
	}
	g.applyWorkspace(result.workspace)
	if !w.editing || w.report == "" || w.lines[0] != "returned" {
		t.Fatal("conflicting import lost the editable draft")
	}
	data, err = os.ReadFile(path)
	if err != nil || string(data) != "concurrent" {
		t.Fatal("concurrent edit was lost", err)
	}

}
func TestExternalDocumentRoundTrip(t *testing.T) {
	copy, err := newExternalCopy("original", "opening-revision")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(copy.dir)
	path := filepath.Join(copy.dir, "document.txt")
	err = os.WriteFile(path, []byte("external\nchange"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	s, err := copy.read()
	if err != nil || s != "external\nchange" || copy.base != "opening-revision" {
		t.Fatal("incorrect returned draft", err)
	}
	err = os.WriteFile(path, []byte{0xff}, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = copy.read()
	if err == nil {
		t.Fatal("invalid working copy accepted")
	}
}
