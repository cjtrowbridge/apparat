// Handles identity setup and locks for a running node.
// The cluster refresh runs after the identity dialog returns.
package app

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

func (node *Runtime) SetupIdentities(dialog identity.Dialog) {
	node.trustMu.Lock()
	node.trustDialog = dialog
	node.trustMu.Unlock()
	defer node.RequestSync()
	node.Keys.Setup(dialog)
	if node.Personal != nil && node.Keys.State() == "unlocked" {
		node.Keys.SetupHostname(dialog)
		if node.Personal.Setup(dialog, node.personalUnlockDevice()) {
			node.openDiscovery.Store(true)
			return
		}
		pub, binding, err := node.Personal.Credentials(node.Keys.Public())
		if err == nil {
			node.mu.Lock()
			if node.database != nil {
				err = node.database.RememberDevice(context.Background(), pub, binding)
				if err == nil && node.serviceManager != nil {
					node.serviceManager.notify()
				}
			}
			node.mu.Unlock()
			if err != nil {
				node.Personal.Lock()
				identity.NotifyIdentityUnavailable(dialog)
			}
		}
	}
}

func (node *Runtime) LockIdentities() {
	node.trustMu.Lock()
	node.trustDialog = nil
	node.reviewedPins = nil
	node.trustMu.Unlock()
	node.Keys.Lock()
	if node.Personal != nil {
		node.Personal.Lock()
	}
}

func (node *Runtime) TakePairingDiscoveryRequest() bool {
	return node.openDiscovery.Swap(false)
}
