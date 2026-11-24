//go:build gui

// Keeps one provisional transcription in the captured editor until finalization.
// Document persistence, undo and form callbacks receive only the accepted text.
package gui

import (
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/widget"
)

type inlineDictation struct {
	field                       *widget.TextInput
	original, expected, partial string
	anchor, caret               int
	changed                     *event.Event
}

func (v *voiceState) targetField(field *widget.TextInput) {
	v.discardInline()
	v.field = field
	if field == nil {
		return
	}
	v.original = field.GetText()
	v.expected = v.original
	e := v.surface()
	if e != nil {
		v.anchor, v.caret = e.anchor, e.caret
	}
}

func (v *inlineDictation) surface() *documentSurface {
	if v.field == nil {
		return nil
	}
	e, _ := v.field.GetWidget().CustomData.(*documentSurface)
	return e
}

func (v *inlineDictation) revise(text string) {
	if text == v.partial {
		return
	}
	e := v.surface()
	if e != nil {
		if !e.editable() {
			return
		}
		a, b := min(v.anchor, v.caret), max(v.anchor, v.caret)
		if len(v.original)-(b-a)+len(text) > domain.WorkspaceTextLimit {
			return
		}
		e.load(v.original[:a] + text + v.original[b:])
		e.caret = a + len(text)
		e.anchor = e.caret
		e.reveal()
	} else {
		if v.changed == nil {
			v.changed = v.field.ChangedEvent
			v.field.ChangedEvent = &event.Event{}
		}
		oldRunes, newRunes := []rune(v.partial), []rune(text)
		same := 0
		for same < len(oldRunes) && same < len(newRunes) && oldRunes[same] == newRunes[same] {
			same++
		}
		for range oldRunes[same:] {
			v.field.Backspace()
		}
		v.field.Insert(string(newRunes[same:]))
	}
	v.partial = text
	v.expected = v.field.GetText()
}

func (v *inlineDictation) discardInline() {
	if v.field != nil && (v.partial != "" || v.expected != v.original) && v.field.GetText() == v.expected {
		e := v.surface()
		if e != nil {
			e.load(v.original)
			e.anchor, e.caret = v.anchor, v.caret
			e.reveal()
		} else {
			v.field.SetText(v.original)
		}
	}
	if v.changed != nil {
		v.field.ChangedEvent = v.changed
	}
	*v = inlineDictation{}
}

func (g *Game) completeInline(text string) {
	v := &g.voice
	field := v.field
	if field == nil || v.surface() != nil || text == "" || v.changed == nil {
		v.discardInline()
		g.finishUtterance(field, text)
		return
	}
	v.revise(text)
	field.ChangedEvent = v.changed
	v.inlineDictation = inlineDictation{}
	field.ChangedEvent.Fire(&widget.TextInputChangedEventArgs{TextInput: field, InputText: field.GetText()})
}
