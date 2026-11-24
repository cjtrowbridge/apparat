// Resolves chat choices at the project owner or from a user's cluster default replica.
// Global replicas reconcile through authenticated worker clients; session overrides stay independent.
package app

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"strings"
	"sync"
	"time"
)

func (n *Runtime) AvailableChatModels(ctx context.Context) ([]domain.ModelOption, error) {
	rows, err := n.KnownWorkers(ctx)
	if err != nil {
		return nil, err
	}
	return domain.ChatModelCatalog(n.LocalInventory(), n.LocalClusterDevice(), rows, time.Now()), nil
}
func (n *Runtime) modelCluster(ctx context.Context) (string, []domain.ClusterDevice, error) {
	rows, err := n.KnownWorkers(ctx)
	if err != nil {
		return "", nil, err
	}
	cluster := ""
	if n.Personal == nil {
		n.mu.Lock()
		defer n.mu.Unlock()
		if n.database != nil {
			cluster, err = n.database.HostModelCluster(ctx)
		}
		return cluster, nil, err
	}
	for _, w := range rows {
		if cluster != "" && cluster != w.Cluster {
			return "", nil, errors.New("choose one active cluster before editing global defaults")
		}
		cluster = w.Cluster
	}
	return cluster, rows, nil
}
func (n *Runtime) ChatModelSettings(ctx context.Context, project, session string, fallback domain.ModelRef) (domain.ModelSettings, error) {
	host, id, remote := strings.Cut(project, "/")
	if remote && n.Personal == nil {
		return domain.ModelSettings{}, httpapi.ErrWorkspaceLocked
	}
	if remote {
		c, err := n.workspaceClient(ctx, host)
		if err != nil {
			return domain.ModelSettings{}, err
		}
		return c.ModelCommand(ctx, httpapi.ModelCommand{Action: "bind", Project: id, Session: session, Fallback: fallback})
	}
	cluster := ""
	active, workers, err := n.modelCluster(ctx)
	if err != nil {
		return domain.ModelSettings{}, err
	}
	if project == "" {
		cluster = active
	}
	if project == "" && len(workers) > 0 {
		n.mu.Lock()
		if n.database == nil {
			n.mu.Unlock()
			return domain.ModelSettings{}, domain.ErrUnavailable
		}
		previous, readErr := n.database.ChatModels(ctx, cluster, project, session)
		n.mu.Unlock()
		if readErr != nil {
			return previous, readErr
		}
		if !previous.Session.Valid() && (fallback.Valid() || previous.Default.Model.Valid()) {
			// Read the latest reachable heads before binding a new chat. A disconnected
			// installation can still resolve from its durable cluster replica.
			refresh, cancel := context.WithTimeout(ctx, 3*time.Second)
			var pending sync.WaitGroup
			for _, worker := range workers {
				pending.Go(func() {
					c, openErr := n.openSavedWorker(refresh, worker)
					if openErr == nil {
						_, _ = n.syncGlobalModel(refresh, c)
					}
				})
			}
			pending.Wait()
			cancel()
		}
	}
	admin := false
	for _, w := range workers {
		admin = admin || w.Admin
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.database == nil {
		return domain.ModelSettings{}, domain.ErrUnavailable
	}
	s, err := n.database.BindChatModel(ctx, cluster, project, session, fallback)
	s.CanDefault = admin
	return s, err
}
func (n *Runtime) SaveChatModel(ctx context.Context, e domain.ModelEdit) (domain.ModelSettings, error) {
	var s domain.ModelSettings
	host, id, remote := strings.Cut(e.Project, "/")
	if remote && n.Personal == nil {
		return domain.ModelSettings{}, httpapi.ErrWorkspaceLocked
	}
	if remote {
		c, err := n.workspaceClient(ctx, host)
		if err != nil {
			return s, err
		}
		return c.ModelCommand(ctx, httpapi.ModelCommand{Action: "save", Project: id, Session: e.Session, Edit: e})
	}
	cluster, workers, err := n.modelCluster(ctx)
	if err != nil {
		return s, err
	}
	if e.Default {
		authorized := false
		for _, w := range workers {
			c, err := n.workspaceClient(ctx, "worker-"+w.Fingerprint)
			if err != nil {
				continue
			}
			access, err := c.WhoAmI(ctx)
			if err != nil || !access.Admin || access.Cluster != cluster {
				continue
			}
			authorized = true
			if e.Project == "" {
				_, err = n.syncGlobalModel(ctx, c)
				if err != nil {
					return s, err
				}
				e.Proof, err = n.Personal.SignModelDefault(n.Keys, cluster, domain.ModelDefault{Model: e.Model, Revision: e.Expected.Revision + 1, Operation: e.Operation})
				if err != nil {
					return s, err
				}
				result, err := c.ModelCommand(ctx, httpapi.ModelCommand{Action: "save", Session: "global-default", Edit: e})
				if err != nil {
					return s, err
				}
				n.mu.Lock()
				if n.database == nil {
					n.mu.Unlock()
					return s, domain.ErrUnavailable
				}
				s, err = n.database.AcceptGlobalChoice(ctx, cluster, e.Session, result.Default)
				n.mu.Unlock()
				n.RequestSync()
				return s, err
			}
			break
		}
		if !authorized {
			return s, errors.New("a connected cluster administrator is required to change defaults")
		}
	}
	if e.Project != "" {
		cluster = ""
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.database == nil {
		return s, domain.ErrUnavailable
	}
	return n.database.SaveChatModel(ctx, cluster, e)
}

func (n *Runtime) cacheGlobalModel(ctx context.Context, cluster string, d domain.ModelDefault) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.database == nil {
		return domain.ErrUnavailable
	}
	return n.database.MergeGlobalModel(ctx, cluster, d)
}
func (n *Runtime) syncGlobalModel(ctx context.Context, c *httpapi.WorkerClient) (domain.ModelDefault, error) {
	remote, err := c.ModelCommand(ctx, httpapi.ModelCommand{Action: "read"})
	if err != nil {
		return domain.ModelDefault{}, err
	}
	cluster := c.Access.Cluster
	if remote.Default.Model.Valid() {
		err = n.cacheGlobalModel(ctx, cluster, remote.Default)
		if err != nil {
			return remote.Default, err
		}
	}
	n.mu.Lock()
	if n.database == nil {
		n.mu.Unlock()
		return remote.Default, domain.ErrUnavailable
	}
	local, err := n.database.ChatModels(ctx, cluster, "", "")
	n.mu.Unlock()
	if err == nil && local.Default != remote.Default && local.Default.Model.Valid() {
		_, err = c.ModelCommand(ctx, httpapi.ModelCommand{Action: "merge", Replica: local.Default})
	}
	return local.Default, err
}

func (n *Runtime) SetChatSessionModel(ctx context.Context, project, id string, model domain.ModelRef) (domain.ChatSession, error) {
	if !model.Valid() {
		return domain.ChatSession{}, errors.New("invalid model")
	}
	return n.withChat(ctx, project, id, nil, func(s *domain.ChatSession) (bool, error) {
		changed := s.Model != model
		s.Model = model
		return changed, nil
	})
}

func (n *Runtime) initialChatModel(ctx context.Context, project string, scope domain.ChatScope) domain.ModelRef {
	cluster := ""
	if scope == domain.GlobalChat {
		project = ""
		var err error
		cluster, _, err = n.modelCluster(ctx)
		if err != nil {
			return domain.ModelRef{}
		}
	}
	n.mu.Lock()
	var settings domain.ModelSettings
	if n.database != nil {
		settings, _ = n.database.ChatModels(ctx, cluster, project, "")
	}
	n.mu.Unlock()
	if settings.Default.Model.Valid() {
		return settings.Default.Model
	}
	options, _ := n.AvailableChatModels(ctx)
	if len(options) > 0 {
		return options[0].Model
	}
	return domain.ModelRef{}
}
