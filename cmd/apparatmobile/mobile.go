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

func UpdateButtonX(width int, height int) int {
	if game == nil {
		return 0
	}
	return game.UpdateButtonX(width, height)
}

func UpdateButtonY(width int, height int) int {
	if game == nil {
		return 0
	}
	return game.UpdateButtonY(width, height)
}

func UpdateButtonW(width int, height int) int {
	if game == nil {
		return 0
	}
	return game.UpdateButtonW(width, height)
}

func UpdateButtonH(width int, height int) int {
	if game == nil {
		return 0
	}
	return game.UpdateButtonH(width, height)
}

func UpdateButtonVisible(width int, height int) bool {
	if game == nil {
		return false
	}
	return game.UpdateButtonVisible(width, height)
}
