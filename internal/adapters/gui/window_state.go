//go:build gui

// Restores desktop placement in monitor-local logical pixels.
// Coalesces move events and preserves the last normal rectangle while maximized.
package gui

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/hajimehoshi/ebiten/v2"
	"log/slog"
	"time"
)

type geometryStore interface {
	WindowState(context.Context) (domain.WindowState, error)
	SaveWindowState(context.Context, domain.WindowState) error
}
type windowTracker struct {
	store           geometryStore
	latest, saved   domain.WindowState
	changed         time.Time
	pending         chan error
	writing         domain.WindowState
	maximizeOnStart bool
}

func (g *Game) restoreWindow() {
	store, ok := g.runtimeInfo.Queries.(geometryStore)
	if !ok {
		return
	}
	g.window.store = store
	w, err := store.WindowState(context.Background())
	if err != nil {
		slog.Warn("window placement read failed")
		return
	}
	if w.Width <= 0 || w.Height <= 0 {
		return
	}
	for _, monitor := range ebiten.AppendMonitors(nil) {
		if monitor.Name() == w.Monitor {
			ebiten.SetMonitor(monitor)
			break
		}
	}
	monitor := ebiten.Monitor()
	if monitor == nil {
		return
	}
	mw, mh := monitor.Size()
	if mw <= 0 || mh <= 0 {
		return
	}
	w.Width = min(w.Width, mw)
	w.Height = min(w.Height, mh)
	ebiten.SetWindowSize(w.Width, w.Height)
	ebiten.SetWindowPosition(max(0, min(w.X, mw-w.Width)), max(0, min(w.Y, mh-w.Height)))
	g.window.maximizeOnStart = w.Maximized
	g.window.latest = w
	g.window.saved = w
}
func (g *Game) resetWindow() {
	if ebiten.IsWindowMaximized() {
		ebiten.RestoreWindow()
	}
	monitor := ebiten.Monitor()
	if monitor == nil {
		return
	}
	mw, mh := monitor.Size()
	w, h := min(960, mw), min(720, mh)
	ebiten.SetWindowSize(w, h)
	ebiten.SetWindowPosition((mw-w)/2, (mh-h)/2)
}
func (g *Game) trackWindow() {
	t := &g.window
	if t.store == nil || ebiten.IsWindowMinimized() || ebiten.IsFullscreen() {
		return
	}
	if t.maximizeOnStart {
		t.maximizeOnStart = false
		ebiten.MaximizeWindow()
		return
	}
	w := t.latest
	w.Maximized = ebiten.IsWindowMaximized()
	if !w.Maximized {
		w.Width, w.Height = ebiten.WindowSize()
		w.X, w.Y = ebiten.WindowPosition()
		monitor := ebiten.Monitor()
		if monitor != nil {
			w.Monitor = monitor.Name()
		}
	}
	if w.Width <= 0 || w.Height <= 0 {
		return
	}
	if w != t.latest {
		t.latest = w
		t.changed = time.Now()
	}
	select {
	case err := <-t.pending:
		t.pending = nil
		if err == nil {
			t.saved = t.writing
		} else {
			slog.Warn("window placement save failed")
			t.changed = time.Now()
		}
	default:
	}
	if t.pending != nil || w == t.saved || time.Since(t.changed) < 500*time.Millisecond {
		return
	}
	result := make(chan error, 1)
	t.pending = result
	t.writing = w
	store := t.store
	go func() {
		result <- store.SaveWindowState(context.Background(), w)
	}()
}
func (g *Game) flushPlacement() {
	t := &g.window
	if t.pending != nil {
		err := <-t.pending
		if err == nil {
			t.saved = t.writing
		}
		t.pending = nil
	}
	if t.store != nil && t.latest.Width > 0 && t.latest != t.saved {
		err := t.store.SaveWindowState(context.Background(), t.latest)
		if err != nil {
			slog.Warn("final window placement save failed")
		}
	}
	if g.prefs.pending != nil {
		<-g.prefs.pending
	}
}
