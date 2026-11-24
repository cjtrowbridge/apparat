//go:build android && gui

// Connects the native multiline editor to the selected checkout session.
// The Java dialog sends edits only through the scoped application callback.
package apparatmobile

import (
	"encoding/json"
	"sync"
)

type Editor interface {
	CopyText(value string)
	PasteText() string
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

func copyText(text string) {
	editor.RLock()
	defer editor.RUnlock()
	if editor.host != nil {
		editor.host.CopyText(text)
	}
}
func pasteText() string {
	editor.RLock()
	defer editor.RUnlock()
	if editor.host != nil {
		return editor.host.PasteText()
	}
	return ""
}
