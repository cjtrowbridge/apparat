package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/cjtrowbridge/apparat/internal/app"
)

func main() {
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
