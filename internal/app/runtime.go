// Owns node initialization, ordered shutdown and the shared application service.
// Presentation is composed by cmd; one OS lock protects one durable node root.
package app

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/discovery"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/config"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"github.com/cjtrowbridge/apparat/internal/logging"
	"github.com/cjtrowbridge/apparat/internal/platform"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type Mode string

const (
	ModeGUI      Mode = "gui"
	ModeHeadless Mode = "headless"
	ModeAuto     Mode = "auto"
)

type Runtime struct {
	// WorkerEvents is set by the standalone daemon before initialization.
	WorkerEvents     io.Writer
	discoveryWorkers sync.WaitGroup
	Keys             *identity.Session
	Personal         *identity.Personal
	workspaceMu      sync.Mutex
	documents        ProjectDocuments
	mu               sync.Mutex
	lifecycle        sync.Mutex
	mode             Mode
	config           config.Config
	database         *persistence.DB
	lock             *platform.Lock
	lastRun          logging.LastRun
	logger           logging.Logger
	serviceManager   *Services
	drivers          map[string]Probe
	server           *httpapi.Server
	state            string
	cancel           context.CancelFunc
	stopped          chan struct{}
	hook             func(string) error
}

func NewRuntime(mode Mode) (*Runtime, error) {
	return NewRuntimeWithConfig(config.Config{Mode: config.Mode(mode)})
}
func NewRuntimeWithConfig(cfg config.Config) (*Runtime, error) {
	mode := Mode(cfg.Mode)
	if mode == "" || mode == ModeAuto {
		mode = ModeGUI
	}
	if mode != ModeGUI && mode != ModeHeadless {
		return nil, errors.New("unsupported runtime mode")
	}
	if cfg.RootDir == "" {
		loaded, err := config.Load(config.Options{DefaultMode: config.Mode(mode)})
		if err != nil {
			return nil, err
		}
		cfg = loaded
	}
	var personal *identity.Personal
	if mode == ModeGUI {
		personal = identity.NewPersonal(filepath.Join(cfg.RootDir, "user-identity"))
	}
	return &Runtime{Personal: personal, Keys: identity.NewSession(cfg.IdentityDir), mode: mode, config: cfg, state: "stopped", drivers: map[string]Probe{}}, nil
}
func (node *Runtime) SetDrivers(drivers map[string]Probe) {
	node.mu.Lock()
	defer node.mu.Unlock()
	if node.state != "stopped" {
		return
	}
	for name, probe := range drivers {
		node.drivers[name] = probe
	}
}
func (node *Runtime) Initialize(ctx context.Context) (result error) {
	node.lifecycle.Lock()
	defer node.lifecycle.Unlock()
	node.mu.Lock()
	defer node.mu.Unlock()
	if node.state == "ready" {
		return nil
	}
	if node.state != "stopped" {
		return domain.ErrUnavailable
	}
	node.state = "starting"
	// The deferred unwind runs under the lifecycle mutex; no observer can use partial resources.
	defer func() {
		if result != nil {
			node.state = "failed"
			cleanup := node.closeLocked()
			_ = node.RecordLastRun("error", "runtime", "startup_failed", "runtime initialization failed", nil)
			result = errors.Join(result, cleanup)
		}
	}()
	result = platform.MigrateRoot(node.config.RootDir, node.config.LegacyRoots)
	if result != nil {
		return result
	}
	result = config.EnsureDirectories(node.config)
	if result != nil {
		return result
	}
	node.lock, result = platform.Acquire(node.config.RootDir, node.config.BinaryName, string(node.mode))
	if result != nil {
		return result
	}
	result = node.stage("lock")
	if result != nil {
		return result
	}
	node.lastRun, result = logging.StartLastRun(node.config.LastRunPath, map[string]any{"binary": node.config.BinaryName, "mode": node.mode, "pid": os.Getpid()})
	if result != nil {
		return result
	}
	node.logger = logging.New(filepath.Join(node.config.LogsDir, "apparat.jsonl"), 1<<20, 3)
	result = node.logger.Write(logging.Event{Component: "runtime", Event: "startup"})
	if result != nil {
		return result
	}
	result = node.stage("directories_ready")
	if result != nil {
		return result
	}
	if node.config.Restore != "" {
		result = persistence.RestoreBackup(ctx, node.config.Restore, node.config.DatabasePath)
		if result != nil {
			return result
		}
	}
	_, priorDB := os.Lstat(node.config.DatabasePath)
	node.database, result = persistence.Open(ctx, node.config.DatabasePath)
	if result != nil {
		return result
	}
	result = node.stage("database")
	if result != nil {
		return result
	}
	result = node.database.Migrate(ctx)
	if result != nil {
		return result
	}
	result = node.database.RecoverWorkspaceRuns(ctx)
	if result != nil {
		return result
	}
	result = node.stage("migrations_ready")
	if result != nil {
		return result
	}
	if node.mode != ModeGUI && identity.Classify(node.config.IdentityDir) == identity.StatusInconsistent {
		return errors.New("identity files inconsistent; use explicit repair/recovery before startup")
	}
	if node.mode == ModeHeadless && node.WorkerEvents != nil {
		switch node.Keys.State() {
		case "missing":
			if os.IsNotExist(priorDB) {
				result = node.Keys.Install(nil, "", "plain", "", "", true)
			}
		case "plain":
			result = node.Keys.Unlock("")
		}
		if result != nil {
			return result
		}
	}
	result = node.stage("identity")
	if result != nil {
		return result
	}
	node.serviceManager = NewServices(node.database, node.drivers, time.Now, func() (string, error) {
		return domain.NewID(time.Now(), nil)
	})
	if node.config.Demo {
		result = node.serviceManager.EnsureDemo(ctx)
		if result != nil {
			return result
		}
	}
	result = node.stage("repository_ready")
	if result != nil {
		return result
	}
	if node.config.Listen != "" {
		node.server, result = httpapi.Listen(node.config.Listen, node)
		if result != nil {
			return result
		}
	}
	if node.server != nil {
		_ = node.RecordLastRun("info", "http", "listener_ready", "loopback listener started", map[string]any{"address": node.server.Address(), "ready": false})
	}
	result = node.stage("listener")
	if result != nil {
		return result
	}
	workCtx, cancel := context.WithCancel(ctx)
	node.cancel = cancel
	if runtime.GOOS == "linux" && !node.config.SmokeTest && !node.config.Inspect && !node.config.Doctor {
		args := []string{"supervise", node.config.DatabasePath}
		if node.WorkerEvents != nil {
			args = append(args, "headless")
		}
		cmd := discovery.Command(workCtx, args...)
		cmd.Stdout = node.WorkerEvents
		cmd.WaitDelay = 3 * time.Second
		node.discoveryWorkers.Add(1)
		go func() {
			defer node.discoveryWorkers.Done()
			_ = cmd.Run()
		}()
	}
	node.stopped = make(chan struct{})
	go node.supervise(workCtx, node.serviceManager, node.stopped)
	result = node.stage("supervision")
	if result != nil {
		return result
	}
	node.state = "ready"
	return node.stage("ready")
}
func (node *Runtime) stage(name string) error {
	_ = node.RecordLastRun("info", "runtime", name, "runtime stage", nil)
	if node.hook != nil {
		return node.hook(name)
	}
	return nil
}
func (node *Runtime) supervise(ctx context.Context, services *Services, done chan struct{}) {
	defer close(done)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := services.Refresh(ctx)
			if err != nil && ctx.Err() == nil {
				_ = node.RecordLastRun("error", "services", "refresh_failed", "service refresh failed", nil)
			}
		}
	}
}
func (node *Runtime) Start(ctx context.Context) error {
	err := node.Initialize(ctx)
	if err != nil {
		return err
	}
	<-ctx.Done()
	return nil
}
func (node *Runtime) Close() error {
	node.lifecycle.Lock()
	defer node.lifecycle.Unlock()
	node.mu.Lock()
	defer node.mu.Unlock()
	if node.state == "stopped" {
		return nil
	}
	node.state = "stopping"
	err := node.closeLocked()
	node.state = "stopped"
	return err
}
func (node *Runtime) closeLocked() error {
	var result error
	node.LockIdentities()
	if node.cancel != nil {
		node.cancel()
		<-node.stopped
		node.discoveryWorkers.Wait()
		node.cancel = nil
		result = errors.Join(result, node.stage("supervision_stopped"))
	}
	if node.server != nil {
		server := node.server
		node.server = nil
		node.mu.Unlock()
		result = errors.Join(result, server.Close())
		node.mu.Lock()
		result = errors.Join(result, node.stage("listener_stopped"))
	}
	node.serviceManager = nil
	if node.database != nil {
		result = errors.Join(result, node.database.Close())
		node.database = nil
		result = errors.Join(result, node.stage("database_closed"))
	}
	_ = node.RecordLastRun("info", "runtime", "shutdown", "runtime resources released", nil)
	if node.lock != nil {
		result = errors.Join(result, node.stage("lock_releasing"))
		result = errors.Join(result, node.lock.Close())
		node.lock = nil
	}
	return result
}
func (node *Runtime) Mode() Mode {
	return node.mode
}
func (node *Runtime) Config() config.Config {
	return node.config
}
func (node *Runtime) RecordLastRun(level, component, event, message string, fields map[string]any) error {
	return node.lastRun.Write(level, component, event, message, fields)
}
func MillisNow() int64 {
	return time.Now().UTC().UnixMilli()
}

func (node *Runtime) SetupIdentities(dialog identity.Dialog) {
	node.Keys.Setup(dialog)
	if node.Personal != nil && node.Keys.State() == "unlocked" {
		node.Personal.Setup(dialog)
		pub, binding, err := node.Personal.Credentials(node.Keys.Public())
		if err == nil {
			node.mu.Lock()
			if node.database != nil {
				err = node.database.RememberDevice(context.Background(), pub, binding)
			}
			node.mu.Unlock()
			if err != nil {
				node.Personal.Lock()
				identity.NotifyIdentityUnavailable(dialog)
			}
		}
	}
}

func (node *Runtime) LockIdentities() {
	node.Keys.Lock()
	if node.Personal != nil {
		node.Personal.Lock()
	}
}
