// Writes global turns on their deterministic home worker and replicates complete revisions.
// A signed-in user can resume a session from another GUI using its stable ID.
package app

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func globalUser(project string) (string, bool) {
	id, ok := strings.CutPrefix(project, "global/")
	b, err := base64.RawURLEncoding.DecodeString(id)
	return id, ok && err == nil && len(b) == 32
}

func (n *Runtime) withGlobal(ctx context.Context, user, id string, h *domain.ChatHeader, apply func(*domain.ChatSession) (bool, error)) (domain.ChatSession, error) {
	a, err := n.database.GlobalArtifact(ctx, user, id)
	fresh := errors.Is(err, sql.ErrNoRows) && h != nil
	if fresh {
		err = h.Validate()
		if err != nil {
			return a.Session, err
		}
		a = domain.GlobalArtifact{User: user, Session: domain.ChatSession{ChatHeader: *h}}
	} else if err != nil {
		return a.Session, err
	}
	changed := false
	if apply != nil {
		changed, err = apply(&a.Session)
	}
	if err == nil && (changed || fresh) {
		if a.Session.Home == "" && n.inference != nil {
			a.Session.Home = n.inference.self
		}
		err = a.Session.Validate()
		if err == nil {
			a.Version++
			err = n.database.PutGlobalArtifact(ctx, a, true)
			if err == nil && n.inference != nil {
				n.inference.observer.pulse()
			}
		}
	}
	return a.Session, err
}

func (p *inferencePool) globalOperation(ctx context.Context, access persistence.Access, q domain.InferenceCommand) (v domain.InferenceReply, err error) {
	g := q.Global
	if q.Operation == "global-list" {
		v.Headers, err = p.db.GlobalHeaders(ctx, access.User, g.Cursor)
		return
	}
	if !domain.ValidID(g.Session) {
		return v, httpapi.ErrDenied
	}
	if q.Operation == "global-get" {
		a, e := p.db.GlobalArtifact(ctx, access.User, g.Session)
		v.Artifact = &a
		return v, e
	}
	p.mu.RLock()
	writer := p.sessionWriter(ctx, access.User, g.Session, p.roster)
	p.mu.RUnlock()
	if writer != p.self || g.Home != "" && g.Home != writer {
		return v, errors.New("session belongs to another worker")
	}
	project := "global/" + access.User
	start, err := domain.ChatTime(g.Session)
	if err != nil {
		return v, err
	}
	h := domain.ChatHeader{Version: 1, ID: g.Session, Scope: domain.GlobalChat, TerminalHost: g.TerminalHost, Home: writer, Started: start, Description: g.Title, Model: g.Model}
	var initial *domain.ChatHeader
	if q.Operation == "global-post" {
		initial = &h
	}
	s, err := p.runtime.withChat(ctx, project, g.Session, initial, func(s *domain.ChatSession) (bool, error) {
		if q.Operation == "global-terminal-result" {
			if s.TerminalHost != access.Device || !domain.ValidID(g.Turn) || len(g.Input) > 4096 {
				return false, httpapi.ErrDenied
			}
			return s.FinishTerminalWait(g.Turn, g.Input, g.Base == "failed"), nil
		}
		if q.Operation == "global-rename" {
			if s.Description == g.Title {
				return false, nil
			}
			if s.Description != g.Base {
				return false, errors.New("session title changed")
			}
			s.Description = g.Title
			return true, s.Validate()
		}
		if s.TerminalHost == "" && g.TerminalHost != "" {
			s.TerminalHost = g.TerminalHost
		}
		if !g.Model.Valid() {
			return false, errors.New("select a model")
		}
		s.Model = g.Model
		if g.Parent != nil && len(s.Turns) > 0 && s.Turns[len(s.Turns)-1].ID != *g.Parent {
			for _, t := range s.Turns {
				if t.ID == g.Turn {
					return p.runtime.queueChatTurn(ctx, project, access.User, s, g.Turn, g.Input)
				}
			}
			if *g.Parent != "" && !slices.ContainsFunc(s.Turns, func(t domain.ChatTurn) bool {
				return t.ID == *g.Parent
			}) {
				return false, errors.New("waiting for previous turn")
			}
			changed, e := s.Apply(g.Turn, g.Input, false)
			if e == nil && changed {
				s.Turns[len(s.Turns)-1].Stopped = "Another device advanced this conversation. Review the intervening reply before resending."
			}
			return changed, e
		}
		if g.Parent != nil && *g.Parent != "" && len(s.Turns) == 0 {
			return false, errors.New("waiting for previous turn")
		}
		return p.runtime.queueChatTurn(ctx, project, access.User, s, g.Turn, g.Input)
	})
	if err == nil {
		v.Artifact = &domain.GlobalArtifact{User: access.User, Session: domain.ChatSession{ChatHeader: s.ChatHeader}}
	}
	return
}

func (p *inferencePool) receiveGlobal(ctx context.Context, caller string, a *domain.GlobalArtifact) error {
	if a == nil || a.Version <= 0 || a.Session.Scope != domain.GlobalChat || a.Session.Validate() != nil {
		return httpapi.ErrDenied
	}
	p.mu.RLock()
	writer := p.sessionWriter(ctx, a.User, a.Session.ID, p.roster)
	if a.Session.Home == caller && p.table.Member(caller) {
		writer, _ = p.db.SessionHome(ctx, a.User, a.Session.ID, caller)
	}
	p.mu.RUnlock()
	if writer != caller {
		return httpapi.ErrDenied
	}
	if caller == p.self {
		return nil
	}
	existing, err := p.db.GlobalArtifact(ctx, a.User, a.Session.ID)
	if err == nil && existing.Version >= a.Version {
		return nil
	}
	p.mu.RLock()
	peers := p.roster.Peers
	p.mu.RUnlock()
	for _, peer := range peers {
		if peer.ID != caller {
			continue
		}
		c, e := p.peerClient(peer)
		if e != nil {
			return e
		}
		reply, e := httpapi.ReadGlobal(ctx, a.Session.ID, a.User, c.Call)
		if e != nil {
			return e
		}
		if reply.Artifact.Version < a.Version {
			return errors.New("source revision unavailable")
		}
		return p.db.PutGlobalArtifact(ctx, *reply.Artifact, false)
	}
	return httpapi.ErrDenied
}

func (p *inferencePool) publishGlobals(ctx context.Context) {
	rows, err := p.db.PendingGlobals(ctx)
	if err != nil {
		return
	}
	p.mu.RLock()
	roster := p.roster
	p.mu.RUnlock()
	for _, a := range rows {
		if p.sessionWriter(ctx, a.User, a.Session.ID, roster) != p.self {
			continue
		}
		notice := a
		notice.Session.Home = p.self
		notice.Session.Turns = nil
		v := p.fanout(ctx, roster, domain.InferenceCommand{Operation: "global-replica", Artifact: &notice})
		if len(v) == len(roster.Peers) {
			_ = p.db.GlobalDelivered(ctx, a)
		}
	}
}

// The SQLite revision and a retry cursor preserve progress across failed disk writes.
func (n *Runtime) mirrorGlobalMarkdown(ctx context.Context) {
	if n.config.ArtifactsDir == "" {
		return
	}
	rows, err := n.database.MarkdownGlobals(ctx)
	if err != nil {
		return
	}
	for _, a := range rows {
		_, safe := globalUser("global/" + a.User)
		if !safe {
			continue
		}
		root := filepath.Join(n.config.ArtifactsDir, "global", a.User)
		err = os.MkdirAll(root, 0700)
		if err != nil {
			continue
		}
		replace := func(s *domain.ChatSession) (bool, error) {
			*s = a.Session
			return true, nil
		}
		_, err = projectgit.WithChat(root, a.Session.ID, nil, replace)
		if errors.Is(err, os.ErrNotExist) {
			_, err = projectgit.WithChat(root, a.Session.ID, &a.Session.ChatHeader, replace)
		}
		if err == nil {
			_ = n.database.MarkGlobalMarkdown(ctx, a)
		}
	}
}
