//go:build android && gui

// Exposes the Android game and updater bridge.
// Dependencies: github.com/cjtrowbridge/apparat/internal/adapters/gui, github.com/cjtrowbridge/apparat/internal/app, github.com/cjtrowbridge/apparat/internal/config.
package apparatmobile

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/cjtrowbridge/apparat/internal/adapters/gui"
	"github.com/cjtrowbridge/apparat/internal/adapters/mockprovider"
	"github.com/cjtrowbridge/apparat/internal/app"
	"github.com/cjtrowbridge/apparat/internal/config"
	"github.com/hajimehoshi/ebiten/v2/mobile"
)

var game *gui.Game
var node *app.Runtime
var bridgeMu sync.Mutex

func init() {
	cfg, err := config.Load(config.Options{DefaultMode: config.ModeGUI, BinaryName: "apparat"})
	if err != nil {
		slog.Error("parse android config", "error", err)
		return
	}
	// The mobile Phase 8 demonstration seeds only a previously empty service store.
	cfg.Demo = true
	runtime, err := app.NewRuntimeWithConfig(cfg)
	if err != nil {
		slog.Error("create android runtime", "error", err)
		return
	}
	runtime.SetDrivers(mockprovider.Drivers())
	node = runtime
	nodeCache = cfg.CacheDir
	node.SetProjectDocuments(projectDocuments{})
	{
		err := runtime.Initialize(context.Background())
		if err != nil {
			slog.Error("initialize android runtime", "error", err)
			return
		}
	}
	_ = runtime.RecordLastRun("info", "android", "mobile_game_ready", "Ebitengine mobile game registered", map[string]any{"root": cfg.RootDir})
	binaryPath, _ := os.Executable()
	workingDir, _ := os.Getwd()
	game = gui.NewGameWithRuntimeInfo(gui.RuntimeInfo{
		ProjectPicker: projectDocuments{},
		SpeakChoice:   sayChoice, MicrophoneReady: microphoneReady,
		DocumentEditor: projectDocuments{},
		PromptValue:    promptValue,
		WorkingDir:     workingDir,
		RuntimePath:    cfg.RootDir,
		BinaryPath:     binaryPath,
		Queries:        runtime,
	})
	game.SetOnCheckForUpdate(func() bool {
		bridgeMu.Lock()
		current := updater
		bridgeMu.Unlock()
		if current == nil {
			game.SetUpdateStatus(gui.UpdateStatusUnableBridge)
			slog.Warn("android update check requested before updater registration", "reason", gui.UpdateStatusUnableBridge)
			return false
		}
		slog.Info("android update check requested")
		current.CheckForUpdate()
		return true
	})
	game.SetInstalledRelease(func() string {
		bridgeMu.Lock()
		current := updater
		bridgeMu.Unlock()
		if current == nil {
			return ""
		}
		return current.InstalledRelease()
	})
	mobile.SetGame(game)

}

func Ready() bool {
	return game != nil && node != nil && node.Health().Ready
}

// BackgroundARGB shares the GUI palette with the Android window.
func BackgroundARGB() int64 {
	return gui.BackgroundARGB()
}

func ActiveTab() string {
	if game == nil {
		return ""
	}
	return game.ActiveTabID()
}

func ReportUpdateStatus(code string) {
	if game == nil {
		slog.Warn("android update status before game registration", "reason", code)
		return
	}
	game.SetUpdateStatus(gui.UpdateStatusCode(code))
	slog.Info("android update status", "reason", code)
}

type Updater interface {
	InstalledRelease() string
	CheckForUpdate()
}

var updater Updater

func RegisterUpdater(u Updater) {
	bridgeMu.Lock()
	updater = u
	bridgeMu.Unlock()
	slog.Info("android updater registered")
}

// SuspendInput cancels captured input on the next game update after Activity pause.
func SuspendInput() {
	if game != nil {
		game.NativeVoice("cancel")
		game.RequestInputCancellation()
		game.ArcadeSuspend()
	}
}

func ArcadeBack() {
	if game != nil {
		game.ArcadeBack()
	}
}

func WorkspaceBack() {
	if game != nil {
		game.RequestWorkspaceBack()
	}
}

func WorkspaceBackAvailable() bool {
	return game != nil && game.WorkspaceBackAvailable()
}

// Voice forwards controls from the active Android text editor to the local decoder.
func Voice(action string) string {
	return game.NativeVoice(action)
}

type ActionSpeech interface {
	Say(string)
	Microphone() bool
}

var choiceSpeech ActionSpeech

func RegisterActionSpeech(host ActionSpeech) {
	bridgeMu.Lock()
	defer bridgeMu.Unlock()
	choiceSpeech = host
}
func actionSpeechHost() ActionSpeech {
	bridgeMu.Lock()
	defer bridgeMu.Unlock()
	return choiceSpeech
}
func sayChoice(text string) {
	host := actionSpeechHost()
	if host != nil {
		host.Say(text)
	}
}
func microphoneReady() bool {
	host := actionSpeechHost()
	return host != nil && host.Microphone()
}
