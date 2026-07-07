package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

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

	runtime, err := app.NewRuntime(app.ModeGUI)
	if err != nil {
		slog.Error("create runtime", "error", err)
		os.Exit(1)
	}

	if err := runtime.Start(context.Background()); err != nil {
		slog.Error("run apparat", "error", err)
		os.Exit(1)
	}
}

func smokeTest() error {
	runtime, err := app.NewRuntime(app.ModeGUI)
	if err != nil {
		return err
	}
	snapshot := runtime.Snapshot()
	fmt.Printf("apparat smoke ok mode=%s tabs=%s voice=%s\n", runtime.Mode(), strings.Join(snapshot.TabTitles(), ","), snapshot.VoiceState)
	return nil
}
