//go:build gui

// Merges authenticated planning snapshots on the HUD thread while preserving open drafts.
// Durable sends and renames carry the signed-in user and the original stable turn ID.
package gui

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"reflect"
	"slices"
	"time"
)

type planningStore interface {
	PlanningUser() string
	PlanningView(context.Context, string, string) (domain.PlanningView, error)
	SavePlanning(context.Context, string, string, domain.GlobalRequest) error
}
type planningRead struct {
	user, id string
	view     domain.PlanningView
	err      error
}
type planningSync struct {
	user, notice string
	pending      chan planningRead
	next         time.Time
}

func (g *Game) savePlanningOperation(op string, q domain.GlobalRequest) bool {
	store, ok := g.runtimeInfo.Queries.(planningStore)
	if !ok {
		g.planning.sync.notice = "Planning storage is unavailable on this device."
		g.layoutDirty = true
		return false
	}
	p := &g.planning.sync
	user := store.PlanningUser()
	if user != p.user || user == "" {
		p.notice = "Unlock your personal identity to send."
		g.layoutDirty = true
		return false
	}
	err := store.SavePlanning(context.Background(), user, op, q)
	if err != nil {
		p.notice = err.Error()
		g.layoutDirty = true
		return false
	}

	p.pending = nil
	p.next = time.Time{}
	p.notice = ""
	return true
}
func (g *Game) pollPlanning() {
	store, ok := g.runtimeInfo.Queries.(planningStore)
	if !ok {
		return
	}
	p := &g.planning.sync
	user := store.PlanningUser()
	if user != p.user {
		g.detachPlanning()
		g.planning = planningUI{sync: planningSync{user: user}}
		g.globalDraft = ""
		g.newPlanning()
		p = &g.planning.sync
	}
	if user == "" || g.shell.Snapshot().ActiveTab().ID() != hud.TabGlobalSession {
		return
	}
	id := ""
	if g.planning.current != nil && g.planning.page != "history" {
		id = g.planning.current.ID
	}
	if p.pending != nil {
		select {
		case r := <-p.pending:
			p.pending = nil
			if r.user == user && r.id == id && r.err == nil {
				g.mergePlanning(r.view)
			}
			p.next = time.Now().Add(2 * time.Second)
		default:
		}
		return
	}
	if time.Now().Before(p.next) {
		return
	}
	result := make(chan planningRead, 1)
	p.pending = result
	parent := g.projects.ctx
	if parent == nil {
		parent = context.Background()
	}
	g.projects.workers.Go(func() {
		ctx, cancel := context.WithTimeout(parent, 30*time.Second)
		defer cancel()
		v, e := store.PlanningView(ctx, user, id)
		result <- planningRead{user, id, v, e}
	})
}
func (g *Game) mergePlanning(v domain.PlanningView) {
	p := &g.planning
	known := map[string]*planningSession{}
	for _, s := range p.sessions {
		known[s.ID] = s
	}
	for _, h := range v.Headers {
		s := known[h.ID]
		if s == nil {
			s = &planningSession{ChatSession: domain.ChatSession{ChatHeader: h}, scroll: 1, loading: true}
			p.sessions = append(p.sessions, s)
			g.layoutDirty = true
		} else if s.Description != h.Description {
			s.Description = h.Description
			g.layoutDirty = true
		}
	}
	slices.SortStableFunc(p.sessions, func(a, b *planningSession) int {
		return a.Started.Compare(b.Started)
	})
	if p.current != nil && v.Session != nil && p.current.ID == v.Session.ID {
		if p.current.syncReason != v.SyncReason || p.current.loading || !reflect.DeepEqual(p.current.ChatSession, *v.Session) || p.current.acknowledged != v.Acknowledged {
			p.current.syncReason = v.SyncReason
			p.current.loading = false
			p.current.ChatSession = *v.Session
			p.current.acknowledged = v.Acknowledged
			g.layoutDirty = true
		}
	}
}
