//go:build gui

// Routes final PTT transcripts to the original editor or the action catalog.
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
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/speech"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type voiceState struct {
	tap, nativeTap                pttLock
	engine                        *speech.Engine
	nativeMu                      sync.Mutex
	globalRequested               bool
	native, canvas, globalCapture *speech.Session
	inlineDictation
	preview, details, status, confirm string
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
	now := time.Now()
	if v.nativeTap.expired(now) && v.native != nil {
		v.native.Release()
	}
	if strings.HasPrefix(action, "press-") || strings.HasPrefix(action, "release-") {
		edge, source, _ := strings.Cut(action, "-")
		decision := tapWait
		if edge == "press" {
			if v.nativeTap.source != "" || v.native == nil {
				decision = v.nativeTap.press(source, now)
			}
		} else {
			decision = v.nativeTap.release(source, now)
		}
		switch decision {
		case tapBegin:
			action = "start"
		case tapSubmit:
			action = "stop"
		default:
			action = "poll"
		}
	}
	switch action {
	case "slide-lock":
		if v.native != nil && v.nativeTap.source == "pointer" {
			v.nativeTap.locked = true
			v.nativeTap.slideLocked = true
		}
	case "global-prompt":
		v.globalRequested = true
		v.globalCapture = v.native
		v.native = nil
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
		v.nativeTap = pttLock{}
		if v.native != nil {
			v.native.Release()
		}
	case "cancel":
		if v.globalCapture != nil {
			v.globalCapture.Cancel()
			v.globalCapture = nil
		}
		v.nativeTap = pttLock{}
		if v.native != nil {
			v.native.Cancel()
			v.native = nil
		}
	}
	if v.native != nil {
		result = v.native.Poll()
		if result.State == "final" || result.State == "error" || result.State == "cancelled" {
			v.native = nil
			v.nativeTap = pttLock{}
		} else if v.nativeTap.locked {
			result.State = "locked"
		}
	} else {
		v.nativeTap = pttLock{}
	}
	data, _ := json.Marshal(result)
	return string(data)
}
func (g *Game) beginPTT(source string) {
	v := &g.voice
	if v.engine == nil || v.canvas != nil {
		return
	}
	f, _ := g.ui.GetFocusedWidget().(*widget.TextInput)
	g.stopChoiceSpeech()
	if g.runtimeInfo.MicrophoneReady != nil && !g.runtimeInfo.MicrophoneReady() {
		v.preview = "Allow microphone access, then press again."
		return
	}
	s, err := v.engine.Start(g.projects.ctx)
	if err != nil {
		v.preview = err.Error()
		return
	}
	g.shell.StartVoiceCapture(source)
	v.canvas = s
	if f != nil && f == g.globalField {
		f.GetWidget().CustomData.(*documentSurface).ime.Confirm()
		g.globalVoiceFocus = true
	}
	v.targetField(f)
}
func (g *Game) endPTT() {
	g.stopInline()
	g.shell.ReleaseVoiceCapture()
	if g.voice.canvas != nil {
		g.voice.canvas.Release()
	}
}
func (g *Game) cancelPTT() {
	g.planning.voiceSession = nil
	g.voice.tap = pttLock{}
	g.shell.CancelVoiceCapture()
	if g.voice.canvas != nil {
		g.voice.canvas.Cancel()
		g.voice.canvas = nil
	}
	g.interruptInline()
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
	if v.tap.expired(time.Now()) {
		g.endPTT()
	}
	if v.canvas != nil {
		if v.field != nil && (g.ui.GetFocusedWidget() != v.field || v.field.GetText() != v.expected || g.gesture.Owner != gesturePTT && caretMayHaveMoved(v.field)) {
			g.cancelPTT()
			return
		}
		r := v.canvas.Poll()
		v.preview = r.Text
		if v.preview == "" {
			v.preview = r.State
		}
		if v.field != nil && (r.State == "recording" || r.State == "finishing") {
			v.revise(r.Text)
			v.preview = ""
		}
		switch r.State {
		case "final":
			send := g.planning.voiceSession != nil && g.planning.voiceSession == g.planning.current && v.field == g.globalField && (strings.TrimSpace(r.Text) != "" || strings.TrimSpace(v.partial) != "")
			g.planning.voiceSession = nil
			g.completeInline(r.Text)
			if send {
				g.submitPlanning()
			}
			v.preview = ""
			v.canvas = nil
			v.tap = pttLock{}
		case "error", "cancelled":
			g.planning.voiceSession = nil
			g.interruptInline()
			v.tap = pttLock{}
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
func (g *Game) finishUtterance(field *widget.TextInput, text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	if field != nil {
		e, _ := field.GetWidget().CustomData.(*documentSurface)
		if e != nil {
			e.insert(text)
			return
		}
		field.Insert(text)
		return
	}
	g.voiceCommand(text)
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

func (g *Game) cancelSpeechBeforeEditing() {
	if g.voice.canvas == nil || g.voice.field == nil {
		return
	}
	keys := inpututil.AppendJustPressedKeys(nil)
	for _, key := range keys {
		if key != ebiten.KeyControlRight {
			g.cancelPTT()
			return
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || len(inpututil.AppendJustPressedTouchIDs(nil)) > 0 {
		g.cancelPTT()
	}
}
