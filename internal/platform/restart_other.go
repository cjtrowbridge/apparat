//go:build !linux || android

// Keeps native Linux self-update restart outside unsupported platforms.
// Android continues to use its package installer and activity lifecycle.
package platform

import "errors"

func RestartProcess() error {
	return errors.New("in-process update restart is unavailable on this platform")
}
