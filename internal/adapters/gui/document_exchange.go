//go:build gui

// An external editor receives a private plain-text copy, never a checkout path.
// Import returns that copy to the selected project's draft and retains its original revision.
package gui

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/ebitenui/ebitenui/widget"
)

type externalCopy struct{ dir, base string }

func newExternalCopy(body, base string) (externalCopy, error) {
	dir, err := os.MkdirTemp("", "apparat-document-")
	if err != nil {
		return externalCopy{}, err
	}
	copy := externalCopy{dir, base}
	err = os.WriteFile(filepath.Join(dir, "document.txt"), []byte(body), 0600)
	if err != nil {
		os.RemoveAll(dir)
	}
	return copy, err
}
func (c externalCopy) read() (string, error) {
	info, err := os.Lstat(filepath.Join(c.dir, "document.txt"))
	if err != nil || !info.Mode().IsRegular() {
		return "", errors.New("save the working copy as a regular text file, then retry")
	}
	f, err := os.OpenInRoot(c.dir, "document.txt")
	if err != nil {
		return "", err
	}
	defer f.Close()
	info, err = f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return "", errors.New("external editor must save a regular text file")
	}
	data, err := io.ReadAll(io.LimitReader(f, domain.WorkspaceTextLimit+1))
	if err != nil {
		return "", err
	}
	s := string(data)
	if !projectgit.ValidText(s) {
		return "", errors.New("save UTF-8 text up to 64 KiB in the external editor, then retry")
	}
	return s, nil
}
func (g *Game) externalDocumentControls(c *widget.Container, w *workspacePanel) {
	key := [2]string{w.id, w.path}
	copy, opened := g.projects.externalCopies[key]
	if opened {
		g.projectButton(c, "Import external edits", func() {
			g.importDocument(w, key, copy)
		})
		return
	}
	if runtime.GOOS != "linux" {
		return
	}
	launcher, err := exec.LookPath("xdg-open")
	if err != nil {
		return
	}
	g.projectButton(c, "Edit externally…", func() {
		copy, err := newExternalCopy(strings.Join(w.lines, "\n"), w.base)
		if err == nil {
			cmd := exec.Command(launcher, filepath.Join(copy.dir, "document.txt"))
			err = cmd.Start()
			if err == nil {
				go cmd.Wait()
			} else {
				os.RemoveAll(copy.dir)
			}
		}
		if err != nil {
			w.report = "Could not open the external working copy: " + err.Error()
		} else {
			if g.projects.externalCopies == nil {
				g.projects.externalCopies = map[[2]string]externalCopy{}
			}
			g.projects.externalCopies[key] = copy
			w.report = "Save the working copy in the external editor, then import it here."
		}
		g.layoutDirty = true
	})
}
func (g *Game) importDocument(w *workspacePanel, key [2]string, copy externalCopy) {
	if w.readOnly || w.editing || g.projects.pending != nil || key != [2]string{w.id, w.path} {
		return
	}
	body, err := copy.read()
	if err != nil {
		w.report = err.Error()
		g.layoutDirty = true
		return
	}
	g.beginDocumentEdit(w)
	w.base = copy.base
	w.lines = strings.Split(body, "\n")
	w.surface = nil
	if !g.keepEditingDraft(w) {
		return
	}
	delete(g.projects.externalCopies, key)
	os.RemoveAll(copy.dir)
}
func (g *Game) copyDocument(s string) {
	if s == "" {
		return
	}
	g.projects.copiedText = s
	if g.runtimeInfo.CopyText != nil {
		g.runtimeInfo.CopyText(s)
		return
	}
	for _, args := range [][]string{{"wl-copy"}, {"xclip", "-selection", "clipboard", "-in"}, {"xsel", "--clipboard", "--input"}, {"pbcopy"}, {"clip.exe"}} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdin = strings.NewReader(s)
		if cmd.Run() == nil {
			return
		}
	}
}
func (g *Game) pasteDocument() string {
	if g.runtimeInfo.PasteText != nil {
		return g.runtimeInfo.PasteText()
	}
	for _, args := range [][]string{{"wl-paste", "--no-newline"}, {"xclip", "-selection", "clipboard", "-out"}, {"xsel", "--clipboard", "--output"}, {"pbpaste"}} {
		b, err := exec.Command(args[0], args[1:]...).Output()
		if err == nil {
			return string(b)
		}
	}
	return g.projects.copiedText
}
