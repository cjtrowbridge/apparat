package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
		return nil
	}
	if err := config.EnsureDirectories(runtime.config); err != nil {
		return err
	}
	runtime.logger = logging.New(filepath.Join(runtime.config.LogsDir, "apparat.jsonl"), 1<<20, 3)
	if err := runtime.logger.Write(logging.Event{Component: "runtime", Event: "startup", Fields: map[string]any{"mode": runtime.mode, "root": runtime.config.RootDir}}); err != nil {
		return err
	}
	db, err := database.Open(ctx, runtime.config.DatabasePath)
	if err != nil {
		return err
	}
	runtime.database = db
	migrations := []database.Migration{{Version: 1, Name: "phase3_core", SQL: phase3Schema}}
	if err := db.ApplyMigrations(ctx, migrations); err != nil {
		return err
	}
	runtime.cluster = cluster.New(db.SQL)
	if err := runtime.cluster.Init(ctx); err != nil {
		return err
	}
	runtime.messages = messaging.New(db.SQL, 5)
	if err := runtime.messages.Init(ctx); err != nil {
		return err
	}
	runtime.started = true
	return nil
}

func (runtime *Runtime) Close() error {
	if runtime.database != nil {
		return runtime.database.Close()
	}
	return nil
}

func (runtime *Runtime) Doctor(ctx context.Context) Diagnostic {
	diag := Diagnostic{Mode: runtime.mode, RootDir: runtime.config.RootDir, DatabasePath: runtime.config.DatabasePath, IdentityStatus: identity.Classify(runtime.config.IdentityDir), Healthy: true, Message: "ok"}
	if err := runtime.Initialize(ctx); err != nil {
		diag.Healthy = false
		diag.Message = err.Error()
		return diag
	}
	for _, path := range []string{runtime.config.RootDir, runtime.config.LogsDir, runtime.config.IdentityDir, filepath.Dir(runtime.config.DatabasePath)} {
		if _, err := os.Stat(path); err != nil {
			diag.Healthy = false
			diag.Message = err.Error()
			return diag
		}
	}
	return diag
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
