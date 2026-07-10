//go:build android && gui

package apparatmobile

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cjtrowbridge/apparat/internal/adapters/gui"
	"github.com/cjtrowbridge/apparat/internal/app"
	"github.com/cjtrowbridge/apparat/internal/config"
	"github.com/hajimehoshi/ebiten/v2/mobile"
)

var game *gui.Game

func init() {
	cfg, err := config.Load(config.Options{DefaultMode: config.ModeGUI, BinaryName: "apparat"})
	if err != nil {
		slog.Error("parse android config", "error", err)
		return
	}
	runtime, err := app.NewRuntimeWithConfig(cfg)
	if err != nil {
		slog.Error("create android runtime", "error", err)
		return
	}
	if err := runtime.Initialize(context.Background()); err != nil {
		slog.Error("initialize android runtime", "error", err)
		return
	}
	_ = runtime.RecordLastRun("info", "android", "mobile_game_ready", "Ebitengine mobile game registered", map[string]any{"root": cfg.RootDir})
	game = gui.NewGame()
	game.SetOnCheckForUpdate(func() {
		if updater != nil {
			updater.CheckForUpdate()
		}
	})
	mobile.SetGame(game)
	_ = fmt.Sprintf("%p", runtime)
}

func Ready() bool {
	return true
}

func ActiveTab() string {
	if game == nil {
		return ""
	}
	return game.ActiveTabID()
}

type Updater interface {
	CheckForUpdate()
}

var updater Updater

func RegisterUpdater(u Updater) {
	updater = u
}
