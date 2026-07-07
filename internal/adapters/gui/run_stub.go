//go:build !gui

package gui

import "context"

func Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
