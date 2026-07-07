package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/cjtrowbridge/apparat/internal/app"
)

func main() {
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
