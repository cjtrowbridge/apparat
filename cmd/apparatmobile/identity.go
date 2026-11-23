//go:build android && gui

// Runs local identity setup through private Android dialogs away from the UI thread.
// The node runtime owns the unlocked session and clears it when the Activity backgrounds.
package apparatmobile

import (
	"context"
	"sync"
)

type KeyDialogs interface {
	Ask(request string) string
}

var identitySetup sync.Mutex
var currentDialogs KeyDialogs

func SetupKeys(dialogs KeyDialogs) {
	if node == nil || dialogs == nil {
		return
	}
	identitySetup.Lock()
	defer identitySetup.Unlock()
	currentDialogs = dialogs
	node.SetupIdentities(dialogs.Ask)
}

func LockKeys() {
	if node != nil {
		node.LockIdentities()
	}
}

func claimHost(network, host string) {
	identitySetup.Lock()
	dialogs := currentDialogs
	identitySetup.Unlock()
	if dialogs != nil {
		node.ClaimWithDialog(context.Background(), network, host, dialogs.Ask)
	}
}

func editScreenName() {
	identitySetup.Lock()
	dialogs := currentDialogs
	identitySetup.Unlock()
	if dialogs != nil {
		node.EditScreenName(context.Background(), dialogs.Ask)
	}
}

func pairPersonal(network, host string) {
	identitySetup.Lock()
	dialogs := currentDialogs
	identitySetup.Unlock()
	if dialogs != nil {
		node.PairDialog(context.Background(), network, host, dialogs.Ask)
	}
}

func checkWorker(id string) {
	identitySetup.Lock()
	dialogs := currentDialogs
	identitySetup.Unlock()
	if dialogs != nil {
		node.CheckWorker(context.Background(), id, dialogs.Ask)
	}
}
