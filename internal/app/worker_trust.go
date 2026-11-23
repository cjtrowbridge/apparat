// Reviews worker identity changes on the local console before sending login credentials.
// Background checks coalesce alerts. The explicit access check can reopen a declined review.
package app

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"strings"
	"time"
)

func approveTrust(dialog identity.Dialog, title, message string) bool {
	if dialog == nil {
		return false
	}
	const approve = "I verified the fingerprint"
	answer, ok := nativeQuestion(dialog, "trust", title, message, "Keep disconnected", approve)
	return ok && answer == approve
}

func (n *Runtime) reviewReplacement(ctx context.Context, dialog identity.Dialog, old, pin, address string) error {
	rows, err := n.KnownWorkers(ctx)
	if err != nil {
		return err
	}
	for _, w := range rows {
		if w.Fingerprint != old {
			continue
		}
		if !w.Admin {
			nativeQuestion(dialog, "trust", w.DisplayName(), "This worker's identity has changed. A cluster administrator must review it.\nSaved: "+old+"\nPresented: "+pin, "Keep disconnected")
			return errors.New("administrator review required")
		}
		// The original owner's cluster identifier is derived from its signed user key.
		// Other administrators need a live check through an independently pinned worker.
		if w.Cluster != identity.ClaimDigest("apparat/personal-cluster/v1/"+n.Personal.Fingerprint()) {
			confirmed := false
			for _, peer := range rows {
				if peer.Cluster != w.Cluster || peer.Fingerprint == old || peer.Address == address {
					continue
				}
				session, err := n.openSavedWorker(ctx, peer)
				if err == nil && session.Access.Admin && session.Access.Admission == "approved" {
					confirmed = true
					break
				}
			}
			if !confirmed {
				nativeQuestion(dialog, "trust", w.DisplayName(), "Worker identity changed. An independently trusted administrator connection is required. Saved: "+old+"\nPresented: "+pin, "Keep disconnected")
				return errors.New("no independent administrator confirmation available")
			}
		}
		message := "The identity at " + address + " changed. Verify the replacement with apparatd trust-host on the worker console or trusted SSH session.\nSaved: " + old + "\nPresented: " + pin
		if !approveTrust(dialog, w.DisplayName(), message) {
			return errors.New("worker kept disconnected")
		}
		client, err := httpapi.OpenWorker(ctx, address, pin, n.Personal, n.Keys)
		if err != nil {
			return err
		}
		if client.Access.User != n.Personal.Fingerprint() || client.Access.Cluster != w.Cluster {
			return errors.New("replacement did not confirm the same cluster and user")
		}
		n.mu.Lock()
		defer n.mu.Unlock()
		if n.database == nil {
			return domain.ErrUnavailable
		}
		err = n.database.ReplaceWorkerPin(ctx, old, pin, address, client.Access)
		if err == nil {
			n.RequestSync()
		}
		return err
	}
	return errors.New("saved worker unavailable")
}

func (n *Runtime) notifyReplacement(w domain.ClusterDevice, change *httpapi.ChangedWorker) {
	n.trustMu.Lock()
	dialog := n.trustDialog
	if dialog == nil || n.reviewedPins[w.Fingerprint] == change.Presented || !n.claimMu.TryLock() {
		n.trustMu.Unlock()
		return
	}
	if n.reviewedPins == nil {
		n.reviewedPins = map[string]string{}
	}
	n.reviewedPins[w.Fingerprint] = change.Presented
	n.trustMu.Unlock()
	go func() {
		defer n.claimMu.Unlock()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		err := n.reviewReplacement(ctx, dialog, w.Fingerprint, change.Presented, change.Address)
		if err != nil {
			nativeQuestion(dialog, "notice", "Worker disconnected", err.Error())
		}
	}()
}

func (n *Runtime) GitAccounts(ctx context.Context, host string) ([]string, error) {
	if host == "local" {
		return projectgit.SavedAccounts()
	}
	rows, err := n.KnownWorkers(ctx)
	if err != nil {
		return nil, err
	}
	for _, w := range rows {
		if host == "worker-"+w.Fingerprint {
			c, err := n.openSavedWorker(ctx, w)
			if err != nil {
				return nil, err
			}
			names, err := c.GitAccounts(ctx)
			if err != nil {
				return nil, err
			}
			for _, name := range names {
				if strings.TrimSpace(name) == "" || len(name) > 39 || !projectgit.ValidAccount(name) {
					return nil, errors.New("invalid account directory")
				}
			}
			return names, nil
		}
	}
	return nil, errors.New("host account directory unavailable")
}
