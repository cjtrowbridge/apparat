//go:build !gui

package gui

import "context"

func Run(ctx context.Context) error {
	return RunWithRuntimeInfo(ctx, RuntimeInfo{})
}

func RunWithRuntimeInfo(ctx context.Context, _ RuntimeInfo) error {
	<-ctx.Done()
	return ctx.Err()
}
