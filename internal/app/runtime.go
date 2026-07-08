package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/gui"
	"github.com/cjtrowbridge/apparat/internal/cluster"
	"github.com/cjtrowbridge/apparat/internal/config"
	"github.com/cjtrowbridge/apparat/internal/database"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"github.com/cjtrowbridge/apparat/internal/logging"
	"github.com/cjtrowbridge/apparat/internal/messaging"
)

type Mode string

const (
	ModeGUI      Mode = "gui"
	ModeHeadless Mode = "headless"
	ModeAuto     Mode = "auto"
)

type Diagnostic struct {
	Healthy        bool
	Mode           Mode
	RootDir        string
	DatabasePath   string
	IdentityStatus identity.Status
	Message        string
}

type Runtime struct {
	mode     Mode
	config   config.Config
	shell    hud.Shell
	database *database.DB
	logger   logging.Logger
	lastRun  logging.LastRun
	cluster  cluster.Directory
	messages messaging.Store
	started  bool
}

func NewRuntime(mode Mode) (*Runtime, error) {
	return NewRuntimeWithConfig(config.Config{Mode: config.Mode(mode)})
}

func NewRuntimeWithConfig(cfg config.Config) (*Runtime, error) {
	mode := Mode(cfg.Mode)
	if mode == "" || mode == ModeAuto {
		mode = ModeGUI
	}
	switch mode {
	case ModeGUI, ModeHeadless:
	default:
		return nil, fmt.Errorf("unsupported runtime mode %q", mode)
	}
	if cfg.RootDir == "" {
		loaded, err := config.Load(config.Options{DefaultMode: config.Mode(mode)})
		if err != nil {
			return nil, err
		}
		cfg = loaded
		cfg.Mode = config.Mode(mode)
	}
	return &Runtime{mode: mode, config: cfg, shell: hud.NewShell()}, nil
}

func (runtime *Runtime) Start(ctx context.Context) error {
	if err := runtime.Initialize(ctx); err != nil {
		return err
	}
	if runtime.mode == ModeGUI {
		return gui.Run(ctx)
	}
	<-ctx.Done()
	if errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}
	return ctx.Err()
}

func (runtime *Runtime) Initialize(ctx context.Context) error {
	if runtime.started {
		_ = runtime.RecordLastRun("info", "runtime", "initialize_skipped", "runtime already initialized", nil)
		return nil
	}
	if err := config.EnsureDirectories(runtime.config); err != nil {
		return runtime.fail("config", "directories_failed", "runtime directory initialization failed", err)
	}
	lastRun, err := logging.StartLastRun(runtime.config.LastRunPath, map[string]any{
		"binary":       runtime.config.BinaryName,
		"mode":         runtime.mode,
		"root":         runtime.config.RootDir,
		"database":     runtime.config.DatabasePath,
		"os":           goruntime.GOOS,
		"architecture": goruntime.GOARCH,
		"go_version":   goruntime.Version(),
		"pid":          os.Getpid(),
		"flags": map[string]bool{
			"doctor":     runtime.config.Doctor,
			"smoke_test": runtime.config.SmokeTest,
		},
	})
	if err != nil {
		return err
	}
	runtime.lastRun = lastRun
	_ = runtime.RecordLastRun("info", "config", "directories_ready", "runtime directories are ready", map[string]any{
		"root":     runtime.config.RootDir,
		"logs":     runtime.config.LogsDir,
		"identity": runtime.config.IdentityDir,
	})
	runtime.logger = logging.New(filepath.Join(runtime.config.LogsDir, "apparat.jsonl"), 1<<20, 3)
	if err := runtime.logger.Write(logging.Event{Component: "runtime", Event: "startup", Fields: map[string]any{"mode": runtime.mode, "root": runtime.config.RootDir}}); err != nil {
		return runtime.fail("logging", "jsonl_failed", "append-only JSONL logger failed", err)
	}
	_ = runtime.RecordLastRun("info", "logging", "jsonl_ready", "append-only JSONL logger is ready", map[string]any{"path": filepath.Join(runtime.config.LogsDir, "apparat.jsonl")})
	db, err := database.Open(ctx, runtime.config.DatabasePath)
	if err != nil {
		return runtime.fail("database", "open_failed", "SQLite open failed", err)
	}
	runtime.database = db
	_ = runtime.RecordLastRun("info", "database", "open_ready", "SQLite database is open", map[string]any{"path": runtime.config.DatabasePath})
	migrations := []database.Migration{{Version: 1, Name: "phase3_core", SQL: phase3Schema}}
	if err := db.ApplyMigrations(ctx, migrations); err != nil {
		return runtime.fail("database", "migrations_failed", "SQLite migrations failed", err)
	}
	_ = runtime.RecordLastRun("info", "database", "migrations_ready", "SQLite migrations are applied", map[string]any{"count": len(migrations)})
	runtime.cluster = cluster.New(db.SQL)
	if err := runtime.cluster.Init(ctx); err != nil {
		return runtime.fail("cluster", "init_failed", "cluster repository initialization failed", err)
	}
	_ = runtime.RecordLastRun("info", "cluster", "repository_ready", "cluster repository is ready", nil)
	runtime.messages = messaging.New(db.SQL, 5)
	if err := runtime.messages.Init(ctx); err != nil {
		return runtime.fail("messaging", "init_failed", "messaging repository initialization failed", err)
	}
	_ = runtime.RecordLastRun("info", "messaging", "repository_ready", "messaging repository is ready", map[string]any{"retry_limit": 5})
	_ = runtime.RecordLastRun("info", "identity", "status_ready", "identity status classified", map[string]any{"status": identity.Classify(runtime.config.IdentityDir)})
	runtime.started = true
	_ = runtime.RecordLastRun("info", "runtime", "ready", "runtime initialization complete", map[string]any{"mode": runtime.mode})
	return nil
}

func (runtime *Runtime) Close() error {
	_ = runtime.RecordLastRun("info", "runtime", "shutdown", "runtime shutdown requested", nil)
	if runtime.database != nil {
		if err := runtime.database.Close(); err != nil {
			return runtime.fail("database", "close_failed", "SQLite close failed", err)
		}
		_ = runtime.RecordLastRun("info", "database", "closed", "SQLite database closed", nil)
	}
	return nil
}

func (runtime *Runtime) Doctor(ctx context.Context) Diagnostic {
	diag := Diagnostic{Mode: runtime.mode, RootDir: runtime.config.RootDir, DatabasePath: runtime.config.DatabasePath, IdentityStatus: identity.Classify(runtime.config.IdentityDir), Healthy: true, Message: "ok"}
	if err := runtime.Initialize(ctx); err != nil {
		diag.Healthy = false
		diag.Message = err.Error()
		_ = runtime.RecordLastRun("error", "doctor", "failed", "doctor initialization failed", map[string]any{"error": err.Error()})
		return diag
	}
	for _, path := range []string{runtime.config.RootDir, runtime.config.LogsDir, runtime.config.IdentityDir, filepath.Dir(runtime.config.DatabasePath)} {
		if _, err := os.Stat(path); err != nil {
			diag.Healthy = false
			diag.Message = err.Error()
			_ = runtime.RecordLastRun("error", "doctor", "path_failed", "doctor path check failed", map[string]any{"path": path, "error": err.Error()})
			return diag
		}
	}
	_ = runtime.RecordLastRun("info", "doctor", "healthy", "doctor checks passed", map[string]any{"identity": diag.IdentityStatus})
	return diag
}

func (runtime *Runtime) RecordLastRun(level string, component string, event string, message string, fields map[string]any) error {
	return runtime.lastRun.Write(level, component, event, message, fields)
}

func (runtime *Runtime) fail(component string, event string, message string, err error) error {
	_ = runtime.RecordLastRun("error", component, event, message, map[string]any{"error": err.Error()})
	return err
}

func (runtime *Runtime) Mode() Mode             { return runtime.mode }
func (runtime *Runtime) Config() config.Config  { return runtime.config }
func (runtime *Runtime) Snapshot() hud.Snapshot { return runtime.shell.Snapshot() }

const phase3Schema = `
CREATE TABLE IF NOT EXISTS local_runtime (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL,
	updated_at_ms INTEGER NOT NULL
);
`

func MillisNow() int64 { return time.Now().UTC().UnixMilli() }
