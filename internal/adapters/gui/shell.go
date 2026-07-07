package gui

import "github.com/cjtrowbridge/apparat/internal/hud"

type Adapter struct {
	shell hud.Shell
}

func NewAdapter() Adapter {
	return Adapter{shell: hud.NewShell()}
}

func (adapter Adapter) Snapshot() hud.Snapshot {
	return adapter.shell.Snapshot()
}
