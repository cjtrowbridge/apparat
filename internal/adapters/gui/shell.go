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

type DocumentSession interface {
	Apply(action, text string) string
}
type RuntimeInfo struct {
	BrowseLAN       func(bool) string
	PoolTasks       func(string)
	ManageWorker    func(string)
	CheckWorker     func(string)
	PairPersonal    func(string, string)
	ClaimHost       func(string, string)
	EditName        func()
	SpeakChoice     func(string)
	MicrophoneReady func() bool
	PromptValue     func(string, string) (string, bool)
	DocumentEditor  interface {
		EditDocument(string, string, string, DocumentSession) error
	}
	ProjectPicker interface{ ChooseFolder() string }
	WorkingDir    string
	RuntimePath   string
	BinaryPath    string
	Queries       ServiceQueries
}

func NewAdapter() Adapter {
	return Adapter{shell: hud.NewShell()}
}

func (adapter Adapter) Snapshot() hud.Snapshot {
	return adapter.shell.Snapshot()
}
