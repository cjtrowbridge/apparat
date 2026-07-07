package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/cjtrowbridge/apparat/internal/app"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--smoke-test" {
		if err := smokeTest(); err != nil {
			slog.Error("smoke test", "error", err)
			os.Exit(1)
		}
		return
	}

	runtime, err := app.NewRuntime(app.ModeHeadless)
	if err != nil {
		slog.Error("create runtime", "error", err)
		os.Exit(1)
	}

	if err := runtime.Start(context.Background()); err != nil {
		slog.Error("run apparatd", "error", err)
		os.Exit(1)
	}
}

func smokeTest() error {
	runtime, err := app.NewRuntime(app.ModeHeadless)
	if err != nil {
		return err
	}
	fmt.Printf("apparatd smoke ok mode=%s tabs=%d\n", runtime.Mode(), len(runtime.Snapshot().Tabs))
	return nil
}
