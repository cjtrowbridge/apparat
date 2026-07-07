package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/cjtrowbridge/apparat/internal/app"
	"github.com/cjtrowbridge/apparat/internal/config"
)

func main() {
	cfg, err := config.Load(config.Options{Args: os.Args[1:], DefaultMode: config.ModeHeadless})
	if err != nil {
		slog.Error("parse config", "error", err)
		os.Exit(2)
	}
	runtime, err := app.NewRuntimeWithConfig(cfg)
	if err != nil {
		slog.Error("create runtime", "error", err)
		os.Exit(1)
	}
	defer func() { _ = runtime.Close() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.SmokeTest {
		if err := smokeTest(ctx, runtime); err != nil {
			slog.Error("smoke test", "error", err)
			os.Exit(1)
		}
		return
	}
	if cfg.Doctor {
		diag := runtime.Doctor(ctx)
		fmt.Printf("apparatd doctor healthy=%t mode=%s root=%s identity=%s message=%s\n", diag.Healthy, diag.Mode, diag.RootDir, diag.IdentityStatus, diag.Message)
		if !diag.Healthy {
			os.Exit(1)
		}
		return
	}

	if err := runtime.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("run apparatd", "error", err)
		os.Exit(1)
	}
}

func smokeTest(ctx context.Context, runtime *app.Runtime) error {
	if err := runtime.Initialize(ctx); err != nil {
		return err
	}
	diag := runtime.Doctor(ctx)
	fmt.Printf("apparatd smoke ok mode=%s tabs=%d root=%s identity=%s\n", runtime.Mode(), len(runtime.Snapshot().Tabs), diag.RootDir, diag.IdentityStatus)
	return nil
}
