// Composes the shared node with explicit mock drivers and process lifecycle.
// Errors return through deferred cleanup before the process exit code is applied.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/adapters/discovery"
	"github.com/cjtrowbridge/apparat/internal/adapters/mockprovider"
	"github.com/cjtrowbridge/apparat/internal/app"
	"github.com/cjtrowbridge/apparat/internal/config"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	os.Exit(run())
}
func run() (code int) {
	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "api-") {
		mode := strings.TrimPrefix(os.Args[1], "api-")
		if mode != "list" && mode != "grant" && mode != "disable" && mode != "revoke" {
			return 2
		}
		needsRule := mode == "grant" || mode == "revoke"
		if (needsRule && len(os.Args) != 3) || (!needsRule && len(os.Args) != 2) {
			return 2
		}
		cfg, err := config.Load(config.Options{DefaultMode: config.ModeHeadless})
		if err != nil {
			return 1
		}
		cmd := discovery.Command(context.Background(), append([]string{strings.TrimPrefix(os.Args[1], "api-"), cfg.DatabasePath}, os.Args[2:]...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if cmd.Run() != nil {
			return 1
		}
		return 0
	}
	cfg, err := config.Load(config.Options{Args: os.Args[1:], DefaultMode: config.ModeHeadless, BinaryName: "apparatd"})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	node, err := app.NewRuntimeWithConfig(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	node.SetDrivers(mockprovider.Drivers())
	defer func() {
		closeErr := node.Close()
		if closeErr != nil {
			fmt.Fprintln(os.Stderr, "shutdown failed:", closeErr)
			code = 1
		}
	}()
	defer func() {
		if recover() != nil {
			_ = node.RecordLastRun("error", "process", "panic", "process panic", nil)
			code = 1
		}
	}()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	err = node.Initialize(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "startup failed:", err)
		return 1
	}
	if cfg.Restore != "" {
		fmt.Println("backup restored to new root", cfg.RootDir)
		return 0
	}
	if cfg.Backup != "" {
		err = node.Backup(ctx, cfg.Backup)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Println("backup verified", cfg.Backup)
		return 0
	}
	if cfg.Inspect || cfg.SmokeTest {
		list, queryErr := node.List(ctx)
		if queryErr != nil {
			fmt.Fprintln(os.Stderr, queryErr)
			return 1
		}
		if cfg.SmokeTest {
			fmt.Printf("apparatd smoke ok mode=%s services=%d root=%s last_run=%s\n", node.Mode(), len(list), cfg.RootDir, cfg.LastRunPath)
		}
		err = json.NewEncoder(os.Stdout).Encode(list)
		if err != nil {
			return 1
		}
		return 0
	}
	if cfg.Doctor {
		diag := node.Doctor(ctx)
		_ = json.NewEncoder(os.Stdout).Encode(diag)
		if !diag.Healthy {
			return 1
		}
		return 0
	}
	fmt.Fprintf(os.Stderr, "apparatd ready root=%s\n", cfg.RootDir)
	err = node.Start(ctx)
	if err != nil && ctx.Err() == nil {
		fmt.Fprintln(os.Stderr, "run failed:", err)
		return 1
	}
	return 0
}
