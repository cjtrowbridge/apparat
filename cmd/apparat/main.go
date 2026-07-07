package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/cjtrowbridge/apparat/internal/app"
	"github.com/cjtrowbridge/apparat/internal/config"
)

func main() {
	cfg, err := config.Load(config.Options{Args: os.Args[1:], DefaultMode: config.ModeGUI})
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
		fmt.Printf("apparat doctor healthy=%t mode=%s root=%s identity=%s message=%s\n", diag.Healthy, diag.Mode, diag.RootDir, diag.IdentityStatus, diag.Message)
		if !diag.Healthy {
			os.Exit(1)
		}
		return
	}

	if err := runtime.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("run apparat", "error", err)
		os.Exit(1)
	}
}

func smokeTest(ctx context.Context, runtime *app.Runtime) error {
	if err := runtime.Initialize(ctx); err != nil {
		return err
	}
	snapshot := runtime.Snapshot()
	diag := runtime.Doctor(ctx)
	fmt.Printf("apparat smoke ok mode=%s tabs=%s voice=%s root=%s identity=%s\n", runtime.Mode(), strings.Join(snapshot.TabTitles(), ","), snapshot.VoiceState, diag.RootDir, diag.IdentityStatus)
	return nil
}
