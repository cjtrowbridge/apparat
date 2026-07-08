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
	cfg, err := config.Load(config.Options{Args: os.Args[1:], DefaultMode: config.ModeHeadless, BinaryName: "apparatd"})
	if err != nil {
		slog.Error("parse config", "error", err)
		os.Exit(2)
	}
	runtime, err := app.NewRuntimeWithConfig(cfg)
	if err != nil {
		slog.Error("create runtime", "error", err)
		os.Exit(1)
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = runtime.RecordLastRun("error", "process", "panic", "panic before process exit", map[string]any{"panic": fmt.Sprint(recovered)})
			panic(recovered)
		}
	}()
	defer func() { _ = runtime.Close() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.SmokeTest {
		if err := smokeTest(ctx, runtime); err != nil {
			_ = runtime.RecordLastRun("error", "process", "smoke_failed", "smoke test failed", map[string]any{"error": err.Error()})
			slog.Error("smoke test", "error", err)
			os.Exit(1)
		}
		_ = runtime.RecordLastRun("info", "process", "clean_exit", "smoke test completed", nil)
		return
	}
	if cfg.Doctor {
		diag := runtime.Doctor(ctx)
		fmt.Printf("apparatd doctor healthy=%t mode=%s root=%s identity=%s message=%s\n", diag.Healthy, diag.Mode, diag.RootDir, diag.IdentityStatus, diag.Message)
		if !diag.Healthy {
			_ = runtime.RecordLastRun("error", "process", "doctor_failed", "doctor command failed", map[string]any{"message": diag.Message})
			os.Exit(1)
		}
		_ = runtime.RecordLastRun("info", "process", "clean_exit", "doctor command completed", nil)
		return
	}

	if err := runtime.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		_ = runtime.RecordLastRun("error", "process", "run_failed", "headless runtime failed", map[string]any{"error": err.Error()})
		slog.Error("run apparatd", "error", err)
		os.Exit(1)
	}
	_ = runtime.RecordLastRun("info", "process", "clean_exit", "headless runtime exited cleanly", nil)
}

func smokeTest(ctx context.Context, runtime *app.Runtime) error {
	if err := runtime.Initialize(ctx); err != nil {
		return err
	}
	diag := runtime.Doctor(ctx)
	fmt.Printf("apparatd smoke ok mode=%s tabs=%d root=%s identity=%s\n", runtime.Mode(), len(runtime.Snapshot().Tabs), diag.RootDir, diag.IdentityStatus)
	return nil
}
