//go:build android && gui

// Connects the native multiline editor to the selected checkout session.
// The Java dialog sends edits only through the scoped application callback.
package apparatmobile

import (
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/gui"
	"sync"
)

type EditorSession interface {
	Apply(action, text string) string
}
type Editor interface {
	Prompt(title, value string) string
	Open(project, title, text string, session EditorSession) string
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
func (projectDocuments) EditDocument(project, title, text string, session gui.DocumentSession) error {
	editor.RLock()
	host := editor.host
	editor.RUnlock()
	if host == nil {
		return errors.New("Android editor unavailable")
	}
	reason := host.Open(project, title, text, session)
	if reason != "" {
		return errors.New(reason)
	}
	return nil
}

func promptValue(title, value string) (string, bool) {
	editor.RLock()
	host := editor.host
	editor.RUnlock()
	if host == nil {
		return "", false
	}
	var result *string
	err := json.Unmarshal([]byte(host.Prompt(title, value)), &result)
	if err != nil || result == nil {
		return "", false
	}
	return *result, true
}
