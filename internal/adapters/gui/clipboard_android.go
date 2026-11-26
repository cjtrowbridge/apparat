//go:build gui && android

// The Activity supplies the clipboard through RuntimeInfo.
// Without an attached Activity there is no foreground clipboard owner.
package gui

import "errors"

func systemCopy(string) error {
	return errors.New("Copy unavailable. Reopen Apparat to reconnect the Android clipboard.")
}
func systemPaste() (string, error) {
	return "", errors.New("Paste unavailable. Reopen Apparat to reconnect the Android clipboard.")
}
