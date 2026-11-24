// Coordinates tab selection and HUD shell state.
// Dependencies: github.com/cjtrowbridge/apparat/internal/hud.
package gui

import (
	"errors"
	"github.com/cjtrowbridge/apparat/internal/hud"
)

var ErrRestart = errors.New("restart Apparat after update")

type Adapter struct {
	shell hud.Shell
}

type RuntimeInfo struct {
	CopyText        func(string)
	PasteText       func() string
	DeviceName      string
	BrowseLAN       func(bool) string
	CheckWorker     func(string)
	PairPersonal    func(string, string)
	ClaimHost       func(string, string)
	EditName        func()
	SpeakChoice     func(string)
	MicrophoneReady func() bool
	PromptValue     func(string, string, bool) (string, bool)
	ProjectPicker   interface{ ChooseFolder() string }
	WorkingDir      string
	RuntimePath     string
	BinaryPath      string
	Queries         ServiceQueries
}

func NewAdapter() Adapter {
	return Adapter{shell: hud.NewShell()}
}

func (adapter Adapter) Snapshot() hud.Snapshot {
	return adapter.shell.Snapshot()
}
