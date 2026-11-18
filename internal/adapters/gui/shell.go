// Coordinates tab selection and HUD shell state.
// Dependencies: github.com/cjtrowbridge/apparat/internal/hud.
package gui

import "github.com/cjtrowbridge/apparat/internal/hud"

type Adapter struct {
	shell hud.Shell
}

type RuntimeInfo struct {
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
