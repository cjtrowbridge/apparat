//go:build gui

// Connects PTT sessions to their originating editor and Settings to the local model catalog.
// Native dialogs poll the same session protocol as the canvas without blocking the UI.
package gui

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cjtrowbridge/apparat/internal/adapters/speech"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type voiceState struct {
	engine                                      *speech.Engine
	nativeMu                                    sync.Mutex
	native, canvas                              *speech.Session
	field                                       *widget.TextInput
	original, preview, details, status, confirm string
}

func (g *Game) initSpeech() {
	store, ok := g.runtimeInfo.Queries.(speech.Selection)
	if ok && g.runtimeInfo.RuntimePath != "" {
		g.voice.engine = speech.New(g.runtimeInfo.RuntimePath, store)
	}
}
func (g *Game) NativeVoice(action string) string {
	v := &g.voice
	v.nativeMu.Lock()
	defer v.nativeMu.Unlock()
	result := speech.Result{State: "idle"}
	switch action {
	case "start":
		if v.engine == nil {
			result = speech.Result{State: "error", Text: "Speech unavailable"}
			break
		}
		if v.native != nil {
			v.native.Cancel()
		}
		s, err := v.engine.Start(g.projects.ctx)
		v.native = s
		if err != nil {
			result = speech.Result{State: "error", Text: err.Error()}
		}
	case "stop":
		if v.native != nil {
			v.native.Release()
		}
	case "cancel":
		if v.native != nil {
			v.native.Cancel()
			v.native = nil
		}
	}
	if v.native != nil {
		result = v.native.Poll()
		if result.State == "final" || result.State == "error" || result.State == "cancelled" {
			v.native = nil
		}
	}
	data, _ := json.Marshal(result)
	return string(data)
}
func (g *Game) beginPTT(source string) {
	g.shell.StartVoiceCapture(source)
	v := &g.voice
	if v.engine == nil || v.canvas != nil {
		return
	}
	f, ok := g.ui.GetFocusedWidget().(*widget.TextInput)
	if !ok {
		v.preview = "Select a text field or open a document to dictate"
		return
	}
	s, err := v.engine.Start(g.projects.ctx)
	if err != nil {
		v.preview = err.Error()
		return
	}
	v.canvas = s
	v.field = f
	v.original = f.GetText()
}
func (g *Game) endPTT() {
	g.shell.ReleaseVoiceCapture()
	if g.voice.canvas != nil {
		g.voice.canvas.Release()
	}
}
func (g *Game) cancelPTT() {
	g.shell.CancelVoiceCapture()
	if g.voice.canvas != nil {
		g.voice.canvas.Cancel()
		g.voice.canvas = nil
	}
	g.voice.field = nil
	g.voice.preview = ""
}
func caretMayHaveMoved(f *widget.TextInput) bool {
	for _, k := range []ebiten.Key{ebiten.KeyArrowLeft, ebiten.KeyArrowRight, ebiten.KeyHome, ebiten.KeyEnd, ebiten.KeyA} {
		if inpututil.IsKeyJustPressed(k) {
			return true
		}
	}
	x, y := ebiten.CursorPosition()
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && image.Pt(x, y).In(f.GetWidget().Rect)
}
func (g *Game) pollSpeech() {
	v := &g.voice
	if v.canvas != nil {
		if g.ui.GetFocusedWidget() != v.field || v.field.GetText() != v.original || caretMayHaveMoved(v.field) {
			g.cancelPTT()
			return
		}
		r := v.canvas.Poll()
		v.preview = r.Text
		if v.preview == "" {
			v.preview = r.State
		}
		switch r.State {
		case "final":
			if r.Text != "" {
				v.field.Insert(r.Text)
			}
			v.preview = ""
			v.canvas = nil
		case "error", "cancelled":
			v.canvas = nil
		}
	}
	if v.engine != nil && g.shell.Snapshot().ActiveTab().ID() == hud.TabSettings {
		id, loaded, msg, busy := v.engine.Status()
		status := fmt.Sprintf("%s %t %s %t", id, loaded, msg, busy)
		if status != v.status {
			v.status = status
			g.layoutDirty = true
		}
	}
}
func (g *Game) speechSettings(c *widget.Container) {
	v := &g.voice
	if v.engine == nil {
		return
	}
	active, loaded, msg, busy := v.engine.Status()
	for _, m := range speech.Models {
		if m.ID == active {
			label := "Selected · " + m.Label + " · Loads on next dictation"
			if loaded {
				label = "Loaded · " + m.Label
			}
			c.AddChild(g.detailText(label))
		}
	}
	if msg != "" {
		c.AddChild(g.detailText(msg))
	}
	for _, m := range speech.Models {
		g.workspaceRow(c, m.Label, func() {
			v.details = m.ID
			v.confirm = ""
			g.layoutDirty = true
		})
		if m.ID != v.details {
			continue
		}
		languages := "English"
		if !strings.HasSuffix(m.ID, ".en") {
			languages = strings.TrimSpace(speech.Languages)
		}
		if !busy && m.ID != active {
			label := "Download and use"
			info, err := os.Stat(filepath.Join(g.runtimeInfo.RuntimePath, "data", "whisper", m.File))
			if m.ID == speech.Default || (err == nil && info.Size() == m.Bytes) {
				label = "Use model"
			}
			if label == "Download and use" && v.confirm != m.ID {
				g.projectButton(c, "Download options…", func() {
					v.confirm = m.ID
					g.layoutDirty = true
				})
			} else {
				if v.confirm == m.ID {
					c.AddChild(g.detailText(fmt.Sprintf("Download %.1f MB to this device? Apparat verifies the model, loads it for push-to-talk, and remembers it for your next session.", float64(m.Bytes)/1e6)))
					g.projectButton(c, "Cancel", func() {
						v.confirm = ""
						g.layoutDirty = true
					})
				}
				g.projectButton(c, label, func() {
					v.confirm = ""
					v.engine.Select(g.projects.ctx, m.ID)
					g.layoutDirty = true
				})
			}
		}
		c.AddChild(g.detailText(fmt.Sprintf("%.1f MB · %s", float64(m.Bytes)/1e6, languages)))
	}
}
