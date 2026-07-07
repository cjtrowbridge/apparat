package app

import (
	"context"
	"fmt"
)

type Mode string

const (
	ModeGUI      Mode = "gui"
	ModeHeadless Mode = "headless"
)

type Runtime struct {
	mode Mode
}

func NewRuntime(mode Mode) (*Runtime, error) {
	switch mode {
	case ModeGUI, ModeHeadless:
		return &Runtime{mode: mode}, nil
	default:
		return nil, fmt.Errorf("unsupported runtime mode %q", mode)
	}
}

func (runtime *Runtime) Start(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (runtime *Runtime) Mode() Mode {
	return runtime.mode
}
