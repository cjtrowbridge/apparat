// Accepts an authenticated turn on any worker, then delivers it to the session home.
// The receipt is the durable queue entry during membership changes or writer downtime.
package app

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

func (p *inferencePool) sessionWriter(ctx context.Context, user, id string, r domain.InferenceRoster) string {
	home, err := p.db.SessionHome(ctx, user, id, "")
	if err == nil {
		return home
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return ""
	}
	return r.GlobalWriter(user, id)
}

func (p *inferencePool) receiveTurn(ctx context.Context, a persistence.Access, g domain.GlobalRequest) (v domain.InferenceReply, err error) {
	if a.Device != "" && g.TerminalHost != "" && g.TerminalHost != a.Device {
		return v, httpapi.ErrDenied
	}
	if !domain.ValidID(g.Session) || !domain.ValidID(g.Turn) || !g.Model.Valid() || strings.TrimSpace(g.Input) == "" || len(g.Home) != 43 {
		return v, errors.New("invalid global turn")
	}
	at, err := domain.ChatTime(g.Session)
	if err != nil {
		return v, err
	}
	header := domain.ChatHeader{Version: 1, Scope: domain.GlobalChat, ID: g.Session, Description: g.Title, Started: at}
	err = header.Validate()
	if err != nil {
		return v, err
	}
	allowed, err := p.db.InferenceAccess(ctx, a.Cluster, a.User, "Chat Pool", "", false)
	if err != nil || !allowed {
		return v, httpapi.ErrDenied
	}
	p.mu.RLock()
	known := p.table != nil && p.table.Member(g.Home)
	p.mu.RUnlock()
	if !known && g.Home != p.self {
		return v, errors.New("unknown session home")
	}
	home, err := p.db.SessionHome(ctx, a.User, g.Session, g.Home)
	if err != nil || home != g.Home {
		return v, errors.New("conflicting session home")
	}
	err = p.db.ReceiveInference(ctx, a.User, g)
	if err == nil {
		v.State = "received"
		p.observer.pulse()
	}
	return
}

func (p *inferencePool) forwardedTurn(ctx context.Context, user string, g domain.GlobalRequest) (v domain.InferenceReply, err error) {
	p.mu.RLock()
	cluster := p.roster.Cluster
	p.mu.RUnlock()
	if g.Home != p.self {
		return v, httpapi.ErrDenied
	}
	a := persistence.Access{User: user, Cluster: cluster}
	_, err = p.receiveTurn(ctx, a, g)
	if err != nil {
		return v, err
	}
	_, err = p.globalOperation(ctx, a, domain.InferenceCommand{Operation: "global-post", Global: g})
	if err == nil {
		v.State = "applied"
		err = p.db.AppliedInferenceReceipt(ctx, user, g.Session, g.Turn)
	}
	return
}

func (p *inferencePool) applyReceipts(ctx context.Context) {
	rows, err := p.db.PendingInferenceReceipts(ctx)
	if err != nil {
		return
	}
	blocked := map[[2]string]bool{}
	for _, row := range rows {
		g := row.Request
		key := [2]string{row.User, g.Session}
		if blocked[key] {
			continue
		}
		if g.Home == p.self {
			_, err = p.forwardedTurn(ctx, row.User, g)
		} else {
			err = p.relayTurn(ctx, row)
		}
		blocked[key] = err != nil
	}
}

func (p *inferencePool) relayTurn(ctx context.Context, r persistence.InferenceReceipt) error {
	p.mu.RLock()
	peers := p.roster.Peers
	p.mu.RUnlock()
	for _, peer := range peers {
		if peer.ID != r.Request.Home {
			continue
		}
		c, err := p.peerClient(peer)
		if err != nil {
			return err
		}
		v, err := c.Call(ctx, domain.InferenceCommand{Operation: "forward-turn", ID: r.User, Global: r.Request})
		if err != nil {
			return err
		}
		if v.State == "applied" {
			return p.db.AppliedInferenceReceipt(ctx, r.User, r.Request.Session, r.Request.Turn)
		}
	}
	return errors.New("session home unavailable")
}
