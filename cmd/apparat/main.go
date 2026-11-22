// Composes the shared node with explicit mock drivers and process lifecycle.
// Errors return through deferred cleanup before the process exit code is applied.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/adapters/gui"
	"github.com/cjtrowbridge/apparat/internal/adapters/keydialog"
	"github.com/cjtrowbridge/apparat/internal/adapters/mockprovider"
	"github.com/cjtrowbridge/apparat/internal/app"
	"github.com/cjtrowbridge/apparat/internal/config"
	"github.com/cjtrowbridge/apparat/internal/platform"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	code := run()
	if code == 75 {
		err := platform.RestartProcess()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Update installed. Close and reopen Apparat to finish restarting.")
			code = 1
		}
	}
	os.Exit(code)
}
func run() (code int) {
	cfg, err := config.Load(config.Options{Args: os.Args[1:], DefaultMode: config.ModeGUI, BinaryName: "apparat"})
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
			fmt.Printf("apparat smoke ok mode=%s services=%d root=%s last_run=%s\n", node.Mode(), len(list), cfg.RootDir, cfg.LastRunPath)
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
	fmt.Fprintf(os.Stderr, "apparat ready root=%s\n", cfg.RootDir)
	if node.Mode() == app.ModeGUI {
		node.Keys.Setup(func(request string) string {
			return keydialog.Run(ctx, request)
		})
		wd, _ := os.Getwd()
		executable, _ := os.Executable()
		err = gui.RunWithRuntimeInfo(ctx, gui.RuntimeInfo{WorkingDir: wd, RuntimePath: cfg.RootDir, BinaryPath: executable, Queries: node})
	} else {
		err = node.Start(ctx)
	}
	if errors.Is(err, gui.ErrRestart) {
		return 75
	}
	if err != nil && ctx.Err() == nil {
		fmt.Fprintln(os.Stderr, "run failed:", err)
		return 1
	}
	return 0
}
