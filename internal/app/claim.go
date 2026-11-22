// Coordinates the local confirmation, scoped TLS request and durable worker registration.
// Untrusted discovery names are labels. The private invitation supplies the trust anchor.
package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/discovery"

	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
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
	status, err := node.authenticateWorker(ctx, network, host)
	if err != nil {
		nativeQuestion(dialog, "notice", "Claim worker", err.Error())
		return
	}
	nativeQuestion(dialog, "notice", host, status)
}

func (node *Runtime) authenticateWorker(ctx context.Context, network, host string) (string, error) {
	_, _, err := node.Personal.Credentials(node.Keys.Public())
	if err != nil {
		return "", errors.New("Unlock your saved Apparat identity to connect.")
	}
	address, err := discovery.ResolveWorker(ctx, network, host)
	if err != nil {
		return "", err
	}
	node.mu.Lock()
	db := node.database
	node.mu.Unlock()
	if db == nil {
		return "", errors.New("local database unavailable")
	}
	pin, err := db.WorkerPin(ctx, host)
	if err != nil {
		return "", err
	}
	client, err := httpapi.OpenWorker(ctx, address, pin, node.Personal, node.Keys)
	if err != nil {
		return "", err
	}
	err = db.SaveWorkerAccess(ctx, host, address, client.Worker, client.Access)
	if err != nil {
		return "", err
	}
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
		}
		node.mu.Unlock()
	}
	if err != nil {
		nativeQuestion(dialog, "notice", "Personal identity", err.Error())
	}
}
