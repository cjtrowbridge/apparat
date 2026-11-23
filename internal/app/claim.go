// Coordinates the local confirmation, scoped TLS request and durable worker registration.
// Discovery supplies an address. The user verifies the fingerprint through the worker console.
package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/discovery"

	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

func (node *Runtime) ClaimWithDialog(ctx context.Context, network, host string, dialog identity.Dialog) {
	if !node.claimMu.TryLock() {
		return
	}
	defer node.claimMu.Unlock()

	if node.Personal == nil {
		return
	}
	if !node.Keys.SetupHostname(dialog) {
		return
	}
	status, err := node.authenticateWorker(ctx, network, host, dialog)
	if err != nil {
		nativeQuestion(dialog, "notice", "Claim worker", err.Error())
		return
	}
	label := domain.PoolDevice{Host: host, Worker: true}
	nativeQuestion(dialog, "notice", label.DisplayName(host), status)
}

func (node *Runtime) authenticateWorker(ctx context.Context, network, host string, dialog identity.Dialog) (string, error) {
	_, _, err := node.Personal.Credentials(node.Keys.Public())
	if err != nil {
		return "", errors.New("Unlock your saved Apparat identity to connect.")
	}
	address, err := node.resolvePeer(ctx, network, host)
	if err != nil {
		return "", err
	}
	node.mu.Lock()
	db := node.database
	node.mu.Unlock()
	if db == nil {
		return "", errors.New("local database unavailable")
	}
	pin, err := db.WorkerPin(ctx, host, address)
	if err != nil {
		return "", err
	}
	observed, err := httpapi.ProbeWorker(ctx, address)
	if err != nil {
		return "", err
	}
	if pin != "" && pin != observed {
		return "Updated the approved worker identity.", node.reviewReplacement(ctx, dialog, pin, observed, address)
	}
	if pin == "" {
		message := "On " + host + ", run apparatd trust-host and compare its fingerprint with:\n" + observed + "\nFor initial setup, authorize your personal identity on the worker with:\napparatd trust-host " + node.Personal.Fingerprint()
		if !approveTrust(dialog, "Authenticate "+host, message) {
			return "", errors.New("worker kept disconnected")
		}
		pin = observed
	}
	client, err := httpapi.OpenWorker(ctx, address, pin, node.Personal, node.Keys)
	if err != nil {
		return "", err
	}
	err = db.SaveWorkerAccess(ctx, host, address, client.Worker, client.Access)
	if err != nil {
		return "", err
	}
	node.mu.Lock()
	if node.serviceManager != nil {
		node.serviceManager.notify()
	}
	node.mu.Unlock()
	node.RequestSync()
	if client.Access.Admin {
		return "Connected as a cluster administrator.", nil
	}
	if client.Access.Admission == "approved" {
		return "Connected with your approved user identity.", nil
	}
	return "Your user identity is awaiting administrator approval.", nil
}

func nativeQuestion(dialog identity.Dialog, op, title, message string, options ...string) (string, bool) {
	body, _ := json.Marshal(map[string]any{"Op": op, "Title": title, "Message": message, "Options": options})
	var answer *string
	err := json.Unmarshal([]byte(dialog(string(body))), &answer)
	if err != nil || answer == nil {
		return "", false
	}
	return *answer, true
}

func (node *Runtime) EditScreenName(ctx context.Context, dialog identity.Dialog) {
	if node.Personal == nil || !node.claimMu.TryLock() {
		return
	}
	defer node.claimMu.Unlock()
	profile, _, err := node.Personal.Credentials(node.Keys.Public())
	if err != nil {
		nativeQuestion(dialog, "notice", "Personal identity", "Unlock your personal and device identities first.")
		return
	}
	name, ok := nativeQuestion(dialog, "text", "Edit screen name", "Current screen name: "+profile.Profile.Name)
	if !ok {
		return
	}
	err = node.Personal.Rename(name)
	if err == nil {
		var binding identity.SignedBinding
		profile, binding, err = node.Personal.Credentials(node.Keys.Public())
		node.mu.Lock()
		if err == nil && node.database != nil {
			err = node.database.RememberDevice(ctx, profile, binding)
			if err == nil && node.serviceManager != nil {
				node.serviceManager.notify()
			}
		}
		node.mu.Unlock()
	}
	if err != nil {
		nativeQuestion(dialog, "notice", "Personal identity", err.Error())
	}
}

func (node *Runtime) KnownWorkers(ctx context.Context) ([]domain.ClusterDevice, error) {
	node.mu.Lock()
	defer node.mu.Unlock()
	if node.database == nil {
		return nil, domain.ErrUnavailable
	}
	if node.Personal == nil {
		return nil, nil
	}
	return node.database.WorkerSummaries(ctx, node.Personal.Fingerprint(), identity.Fingerprint(node.Keys.Public()))
}

func (n *Runtime) CheckWorker(ctx context.Context, id string, dialog identity.Dialog) {
	rows, err := n.KnownWorkers(ctx)
	if err != nil {
		return
	}
	for _, worker := range rows {
		if worker.Fingerprint != id {
			continue
		}
		client, err := n.openSavedWorker(ctx, worker)
		if err != nil {
			var change *httpapi.ChangedWorker
			if errors.As(err, &change) {
				err = n.reviewReplacement(ctx, dialog, id, change.Presented, change.Address)
				if err != nil {
					nativeQuestion(dialog, "notice", "Worker trust", err.Error())
				}
				return
			}
			nativeQuestion(dialog, "notice", worker.DisplayName(), "Could not authenticate at the saved address. Use discovery to reconnect if the worker moved.")
			return
		}
		access, err := client.WhoAmI(ctx)
		if err != nil || access.Cluster != worker.Cluster || access.User != n.Personal.Fingerprint() {
			nativeQuestion(dialog, "notice", worker.DisplayName(), "The worker did not confirm access for this user and cluster.")
			return
		}
		status := "The worker confirmed your approved user access."
		if access.Admin {
			status = "The worker confirmed your administrator access."
		}
		nativeQuestion(dialog, "notice", worker.DisplayName(), status)
		return
	}
}

func (node *Runtime) resolvePeer(ctx context.Context, network, host string) (string, error) {
	if node.ResolvePeer != nil {
		return node.ResolvePeer(ctx, network, host)
	}
	return discovery.ResolveWorker(ctx, network, host)
}

func (n *Runtime) openSavedWorker(ctx context.Context, worker domain.ClusterDevice) (*httpapi.WorkerClient, error) {
	client, err := httpapi.OpenWorker(ctx, worker.Address, worker.Fingerprint, n.Personal, n.Keys, &n.observations)
	if err == nil {
		if client.Access.Cluster != worker.Cluster || client.Access.User != n.Personal.Fingerprint() {
			return nil, errors.New("unexpected worker cluster or user")
		}
		return client, nil
	}
	var changed *httpapi.ChangedWorker
	if errors.As(err, &changed) || errors.Is(err, httpapi.ErrDenied) {
		return nil, err
	}
	resolver := n.ResolveSaved
	if resolver == nil {
		resolver = discovery.ResolveSaved
	}
	address, err := resolver(ctx, worker.Host, worker.Address)
	if err != nil {
		return nil, err
	}
	client, err = httpapi.OpenWorker(ctx, address, worker.Fingerprint, n.Personal, n.Keys, &n.observations)
	if err != nil {
		return nil, err
	}
	if client.Access.Cluster != worker.Cluster || client.Access.User != n.Personal.Fingerprint() {
		return nil, errors.New("unexpected worker cluster or user")
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.database == nil {
		return nil, domain.ErrUnavailable
	}
	err = n.database.UpdateWorkerAddress(ctx, worker.Fingerprint, client.Access.User, address)
	if err == nil && n.serviceManager != nil {
		n.serviceManager.notify()
	}
	return client, err
}

func (n *Runtime) ScreenName() string {
	if n.Personal == nil {
		return ""
	}
	return n.Personal.ScreenName()
}
