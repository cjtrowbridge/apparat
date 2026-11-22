//go:build android && gui

// Runs local identity setup through private Android dialogs away from the UI thread.
// The node runtime owns the unlocked session and clears it when the Activity backgrounds.
package apparatmobile

import "sync"

type KeyDialogs interface {
	Ask(request string) string
}

var identitySetup sync.Mutex

func SetupKeys(dialogs KeyDialogs) {
	if node == nil || dialogs == nil {
		return
	}
	identitySetup.Lock()
	defer identitySetup.Unlock()
	node.SetupIdentities(dialogs.Ask)
}

func LockKeys() {
	if node != nil {
		node.LockIdentities()
	}
}
