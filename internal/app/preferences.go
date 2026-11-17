// Routes local GUI preferences and placement through the owning runtime.
// Desktop and Android share the same SQLite-backed preference commands.
package app

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

func (n *Runtime) Preferences(ctx context.Context) (domain.Preferences, error) {
	return n.database.Preferences(ctx)
}
func (n *Runtime) SavePreferences(ctx context.Context, p domain.Preferences) error {
	return n.database.SavePreferences(ctx, p)
}
func (n *Runtime) WindowState(ctx context.Context) (domain.WindowState, error) {
	return n.database.WindowState(ctx)
}
func (n *Runtime) SaveWindowState(ctx context.Context, w domain.WindowState) error {
	return n.database.SaveWindowState(ctx, w)
}
