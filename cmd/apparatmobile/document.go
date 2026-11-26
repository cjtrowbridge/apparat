//go:build android && gui

// Connects the native multiline editor to the selected checkout session.
// The Java dialog sends edits only through the scoped application callback.
package apparatmobile

import (
	"encoding/json"
	"errors"
	"sync"
)

type Editor interface {
	CopyText(value string) error
	PasteText() (string, error)
	Prompt(title, value string, singleLine bool) string
}

var editor struct {
	sync.RWMutex
	host Editor
}

func RegisterEditor(host Editor) {
	editor.Lock()
	defer editor.Unlock()
	editor.host = host
}
func promptValue(title, value string, singleLine bool) (string, bool) {
	editor.RLock()
	host := editor.host
	editor.RUnlock()
	if host == nil {
		return "", false
	}
	var result *string
	err := json.Unmarshal([]byte(host.Prompt(title, value, singleLine)), &result)
	if err != nil || result == nil {
		return "", false
	}
	return *result, true
}

func copyText(text string) error {
	editor.RLock()
	host := editor.host
	editor.RUnlock()
	if host != nil {
		return host.CopyText(text)
	}
	return errors.New("Copy unavailable. Reopen Apparat to reconnect the Android clipboard.")
}
func pasteText() (string, error) {
	editor.RLock()
	host := editor.host
	editor.RUnlock()
	if host != nil {
		return host.PasteText()
	}
	return "", errors.New("Paste unavailable. Reopen Apparat to reconnect the Android clipboard.")
}
