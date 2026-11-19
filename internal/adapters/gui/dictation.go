//go:build gui

// Connects PTT sessions to their originating editor and Settings to the local model catalog.
// Native dialogs poll the same session protocol as the canvas without blocking the UI.
package gui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cjtrowbridge/apparat/internal/adapters/speech"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
)

type voiceState struct {
	engine                             *speech.Engine
	nativeMu                           sync.Mutex
	native, canvas                     *speech.Session
	field                              *widget.TextInput
	original, preview, details, status string
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
func (g *Game) pollSpeech() {
	v := &g.voice
	if v.canvas != nil {
		if g.ui.GetFocusedWidget() != v.field || v.field.GetText() != v.original {
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
		id, msg, busy := v.engine.Status()
		status := fmt.Sprintf("%s %s %t", id, msg, busy)
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
	active, msg, busy := v.engine.Status()
	for _, m := range speech.Models {
		if m.ID == active {
			c.AddChild(g.detailText("Dictation model · " + m.Label))
		}
	}
	if msg != "" {
		c.AddChild(g.detailText(msg))
	}
	for _, m := range speech.Models {
		g.workspaceRow(c, m.Label, func() {
			v.details = m.ID
			g.layoutDirty = true
		})
		if m.ID != v.details {
			continue
		}
		languages := "English"
		if !strings.HasSuffix(m.ID, ".en") {
			languages = strings.TrimSpace(speech.Languages)
		}
		c.AddChild(g.detailText(fmt.Sprintf("%.1f MB · %s", float64(m.Bytes)/1e6, languages)))
		if !busy && m.ID != active {
			label := "Download and use"
			info, err := os.Stat(filepath.Join(g.runtimeInfo.RuntimePath, "data", "whisper", m.File))
			if m.ID == speech.Default || (err == nil && info.Size() == m.Bytes) {
				label = "Use model"
			}
			g.projectButton(c, label, func() {
				v.engine.Select(g.projects.ctx, m.ID)
				g.layoutDirty = true
			})
		}
	}
}
