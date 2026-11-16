//go:build !gui

// Reports the GUI-disabled build behavior.
// Uses package-local contracts and Go standard-library types.
package gui

import "context"

func Run(ctx context.Context) error {
	return RunWithRuntimeInfo(ctx, RuntimeInfo{})
}

func RunWithRuntimeInfo(ctx context.Context, _ RuntimeInfo) error {
	<-ctx.Done()
	return ctx.Err()
}
