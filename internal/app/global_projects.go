// Federates the global projects tool across the enrolled workers using the caller's scope.
// Results identify unavailable owners and paginate by the same sortable project IDs as local queries.
package app

import (
	"context"
	"encoding/json"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"slices"
	"strings"
)

func (p *inferencePool) readProjectCatalog(ctx context.Context, caller string, q domain.InferenceCommand) (v domain.InferenceReply, err error) {
	g := q.Global
	p.mu.RLock()
	r := p.roster
	p.mu.RUnlock()
	if !domain.ValidID(g.Session) || p.sessionWriter(ctx, q.ID, g.Session, r) != caller {
		return v, httpapi.ErrDenied
	}
	ok, err := p.db.CatalogMember(ctx, r.Cluster, q.ID)
	if err != nil || !ok {
		return v, httpapi.ErrDenied
	}
	v.Projects, v.Cursor, err = p.db.ChatCatalog(ctx, "", g.Input, g.Cursor)
	return
}
func (p *inferencePool) globalProjects(ctx context.Context, user, session, selector, cursor string) (string, error) {
	p.mu.RLock()
	r := p.roster
	p.mu.RUnlock()
	replies := p.fanout(ctx, r, domain.InferenceCommand{Operation: "catalog", ID: user, Global: domain.GlobalRequest{Session: session, Input: selector, Cursor: cursor}})
	var out struct {
		Projects []domain.ChatProject `json:"projects"`
		Cursor   string               `json:"next_cursor,omitempty"`
		Missing  []string             `json:"unavailable_workers,omitempty"`
	}
	out.Projects = []domain.ChatProject{}
	more := false
	for _, peer := range r.Peers {
		v, ok := replies[peer.ID]
		if !ok {
			out.Missing = append(out.Missing, peer.ID)
			continue
		}
		for _, project := range v.Projects {
			project.Owner = peer.ID
			out.Projects = append(out.Projects, project)
		}
		more = more || v.Cursor != ""
	}
	slices.SortFunc(out.Projects, func(a, b domain.ChatProject) int {
		return strings.Compare(b.ID, a.ID)
	})
	if len(out.Projects) > 32 {
		out.Projects = out.Projects[:32]
		more = true
	}
	if more && len(out.Projects) > 0 {
		out.Cursor = out.Projects[len(out.Projects)-1].ID
	}
	b, err := json.Marshal(out)
	return string(b), err
}
