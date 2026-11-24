//go:build gui

// Drives one asynchronous, scope-addressed model picker on the HUD thread.
// Selection is tentative until a successful save returns to its originating chat.
package gui

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"time"
)

type modelStore interface {
	AvailableChatModels(context.Context) ([]domain.ModelOption, error)
	ChatModelSettings(context.Context, string, string, domain.ModelRef) (domain.ModelSettings, error)
	SaveChatModel(context.Context, domain.ModelEdit) (domain.ModelSettings, error)
}
type modelAddress struct{ project, session string }
type modelReply struct {
	settings domain.ModelSettings
	options  []domain.ModelOption
	err      error
	saved    bool
}
type modelPicker struct {
	expected                domain.ModelDefault
	address                 modelAddress
	settings                domain.ModelSettings
	options                 []domain.ModelOption
	pending                 chan modelReply
	selected                domain.ModelRef
	loaded, changed, saving bool
	next                    time.Time
	notice, op              string
	caret, anchor           int
	restore                 bool
}

func (g *Game) modelAddress() (modelAddress, bool) {
	switch g.shell.Snapshot().ActiveTab().ID() {
	case hud.TabGlobalSession:
		p := &g.planning
		if p.current != nil && (p.page == "" || p.page == "model-picker") {
			return modelAddress{"", p.current.ID}, true
		}
	case hud.TabProjects:
		w := &g.projects.workspace
		if g.inWorkspace() && (w.mode == "conversation" || w.mode == "model-picker") {
			thread := w.thread
			if thread == "" {
				thread = "General"
			}
			return modelAddress{w.id, thread}, true
		}
	}
	return modelAddress{}, false
}
func (g *Game) modelPageOpen() bool {
	if g.shell.Snapshot().ActiveTab().ID() == hud.TabGlobalSession {
		return g.planning.page == "model-picker"
	}
	return g.inWorkspace() && g.projects.workspace.mode == "model-picker"
}
func (g *Game) pollChatModels() {
	p := &g.planning.picker
	addr, active := g.modelAddress()
	if !active {
		return
	}
	store, ok := g.runtimeInfo.Queries.(modelStore)
	if !ok {
		return
	}
	if addr != p.address {
		*p = modelPicker{address: addr}
	}
	if p.pending != nil {
		select {
		case r := <-p.pending:
			p.pending = nil
			p.saving = false
			if r.err != nil {
				p.notice = r.err.Error()
				if errors.Is(r.err, domain.ErrModelConflict) {
					p.op = ""
				}
				p.next = time.Now().Add(time.Second)
				g.layoutDirty = true
				return
			}
			changed := !p.loaded || r.settings != p.settings || (!r.saved && !reflect.DeepEqual(p.options, r.options))
			p.loaded = true
			if r.saved {
				r.settings.CanDefault = p.settings.CanDefault
			}
			p.settings = r.settings
			if !r.saved {
				p.options = r.options
			}
			if !p.changed {
				p.selected = r.settings.Session
			}
			if addr.project == "" {
				g.planning.current.Model = r.settings.Session
			}
			if r.saved && g.modelPageOpen() {
				g.closeModels()
			}
			if changed || r.saved {
				g.layoutDirty = true
			}
		default:
		}
		return
	}
	if time.Now().Before(p.next) {
		return
	}
	p.next = time.Now().Add(4 * time.Second)
	ch := make(chan modelReply, 1)
	p.pending = ch
	ctx := g.runContext
	if ctx == nil {
		ctx = context.Background()
	}
	go func() {
		options, err := store.AvailableChatModels(ctx)
		var s domain.ModelSettings
		if err == nil {
			var first domain.ModelRef
			if len(options) > 0 {
				first = options[0].Model
			}
			s, err = store.ChatModelSettings(ctx, addr.project, addr.session, first)
		}
		ch <- modelReply{settings: s, options: options, err: err}
	}()
}
func (g *Game) openModels() {
	addr, active := g.modelAddress()
	if !active {
		return
	}
	p := &g.planning.picker
	if addr != p.address {
		*p = modelPicker{address: addr}
	}
	p.selected = p.settings.Session
	p.changed = false
	p.notice = ""
	p.op = ""
	p.next = time.Time{}
	if addr.project == "" {
		if g.globalField != nil {
			e := g.globalField.GetWidget().CustomData.(*documentSurface)
			e.ime.Confirm()
			p.caret = e.caret
			p.anchor = e.anchor
			p.restore = true
		}
		g.planningPage("model-picker")
	} else {
		g.projects.workspace.mode = "model-picker"
		g.layoutDirty = true
	}
}
func (g *Game) closeModels() {
	p := &g.planning.picker
	if p.saving {
		return
	}
	if p.address.project == "" {
		g.openPlanning(g.planning.current)
	} else {
		g.projects.workspace.mode = "conversation"
		g.layoutDirty = true
	}
	p.changed = false
	p.notice = ""
	p.op = ""
}
func (p *modelPicker) available(m domain.ModelRef) bool {
	return slices.ContainsFunc(p.options, func(o domain.ModelOption) bool {
		return o.Model == m
	})
}
func (g *Game) chatModelReady() bool {
	_, managed := g.runtimeInfo.Queries.(modelStore)
	if !managed {
		return true
	}
	address, active := g.modelAddress()
	p := &g.planning.picker
	return active && p.address == address && p.settings.Session.Valid()
}
func (g *Game) modelLabel() string {
	p := &g.planning.picker
	m := p.settings.Session
	addr, _ := g.modelAddress()
	if addr != p.address || !m.Valid() {
		return "Choose model…"
	}
	if !p.available(m) {
		return m.Name + " · Unavailable"
	}
	return m.Name
}
func (g *Game) addModelButton(c *widget.Container) {
	width := g.hudPreferredWidth() - 2*pttRadius - pttMargin - pttTravel - 40
	if g.inWorkspace() {
		width = g.detailPanePreferredWidth() - 56
	} else if g.height < 590 {
		width -= 70
	}
	name := g.modelLabel()
	g.projectButton(c, g.planningTitle(name, max(40, width)), g.openModels)
}
func (g *Game) saveModel(asDefault bool) {
	p := &g.planning.picker
	store, ok := g.runtimeInfo.Queries.(modelStore)
	if !ok || p.saving || !p.loaded || !p.available(p.selected) || asDefault && !p.settings.CanDefault {
		return
	}
	if p.op == "" {
		p.expected = p.settings.Default
		p.op, _ = domain.NewID(time.Now(), nil)
	}
	e := domain.ModelEdit{Project: p.address.project, Session: p.address.session, Model: p.selected, Default: asDefault, Expected: p.expected, Operation: p.op}
	ch := make(chan modelReply, 1)
	p.pending = ch
	p.saving = true
	p.notice = "Saving…"
	g.layoutDirty = true
	ctx := g.runContext
	if ctx == nil {
		ctx = context.Background()
	}
	go func() {
		s, err := store.SaveChatModel(ctx, e)
		ch <- modelReply{settings: s, err: err, saved: err == nil}
	}()
}
func (g *Game) modelCards(body *widget.Container) {
	p := &g.planning.picker
	actions, models := g.detailCard(), g.detailCard()
	scope := "global"
	if p.address.project != "" {
		scope = "project"
	}
	save := g.projectButton(actions, "Save for this chat", func() {
		g.saveModel(false)
	})
	save.GetWidget().Disabled = p.saving || !p.available(p.selected)
	def := g.projectButton(actions, "Set as "+scope+" default", func() {
		g.saveModel(true)
	})
	def.GetWidget().Disabled = save.GetWidget().Disabled || !p.settings.CanDefault
	cancel := g.projectButton(actions, "Cancel", g.closeModels)
	cancel.GetWidget().Disabled = p.saving
	if p.loaded && !p.settings.CanDefault {
		actions.AddChild(g.hudText("Only cluster administrators can change defaults."))
	}
	if p.notice != "" {
		actions.AddChild(g.hudText(p.notice))
	}
	models.AddChild(g.hudText("Available models"))
	width := g.detailPanePreferredWidth() - 64
	if g.collapsed() || runtime.GOOS == "android" {
		width = g.hudPreferredWidth() - 64
	}
	list := slices.Clone(p.options)
	for _, m := range []domain.ModelRef{p.settings.Session, p.settings.Default.Model} {
		if m.Valid() && !slices.ContainsFunc(list, func(o domain.ModelOption) bool {
			return o.Model == m
		}) {
			list = append(list, domain.ModelOption{Model: m})
		}
	}
	for _, o := range list {
		label := o.Model.Name + " · " + o.Model.Provider + " · " + strings.Join(o.Nodes, ", ")
		if len(o.Nodes) == 0 {
			label = o.Model.Name + " · Unavailable"
		}
		if o.Model == p.settings.Default.Model {
			label += " · " + scope + " default"
		}
		if o.Model == p.selected {
			label += " · Selected"
		}
		b := g.projectButton(models, wrapSelectorLabel(label, *g.theme.ButtonTheme.TextFace, max(80, width)), func() {
			p.selected = o.Model
			p.changed = true
			p.op = ""
			p.notice = ""
			g.layoutDirty = true
		})
		b.GetWidget().LayoutData = widget.RowLayoutData{Stretch: true}
		b.GetWidget().Disabled = p.saving || len(o.Nodes) == 0
		if o.Model == p.selected {
			img := tabButtonImage()
			img.Idle = img.Pressed
			b.SetImage(img)
		}
	}
	if len(list) == 0 {
		label := "Looking for models…"
		if p.loaded {
			label = "No text-generation models available"
		}
		models.AddChild(g.hudText(label))
	}
	for _, b := range []*widget.Button{save, def, cancel} {
		b.GetWidget().LayoutData = widget.RowLayoutData{Stretch: true}
	}
	body.AddChild(actions, models)
}
func (g *Game) globalModelPage(s *planningSession) *widget.Container {
	content := planningColumn()
	g.modelCards(content)
	scroll := g.planningScroll(content)
	if g.collapsed() || runtime.GOOS == "android" {
		box := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1), widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true}))))
		box.AddChild(scroll)
		return box
	}
	box := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(2), widget.GridLayoutOpts.Stretch([]bool{false, true}, []bool{true}), widget.GridLayoutOpts.Spacing(bodyGap, 0))))
	left := planningColumn()
	left.GetWidget().MinWidth = g.masterListWidth()
	left.AddChild(g.hudText("Current chat"), g.detailText(s.Description))
	box.AddChild(left, scroll)
	return box
}
func (g *Game) modelActions(add registerAction) {
	addr, active := g.modelAddress()
	if !active {
		return
	}
	owner := addr.project + "/" + addr.session
	if !g.modelPageOpen() {
		if addr.project != "" {
			add("chat-model", owner, "choose", "Choose chat model", "Project chat / Choose model", g.openModels)
		}
		return
	}
	p := &g.planning.picker
	if p.saving {
		return
	}
	add("chat-model", owner, "cancel", "Cancel model selection", "Chat / Model / Cancel", g.closeModels)
	if p.available(p.selected) {
		add("chat-model", owner, "save", "Save for this chat", "Chat / Model / Save", func() {
			g.saveModel(false)
		})
		if p.settings.CanDefault {
			add("chat-model", owner, "default", "Set as default", "Chat / Model / Set default", func() {
				g.saveModel(true)
			})
		}
	}
	for _, o := range p.options {
		key, _ := json.Marshal(o.Model)
		add("choose-model", owner, string(key), o.Model.Name, "Chat / Model / "+o.Model.Name+" / "+strings.Join(o.Nodes, ", "), func() {
			p.selected = o.Model
			p.changed = true
			p.op = ""
			g.layoutDirty = true
		})
	}
}
