// Verifies defines configuration values and their defaults.
// Uses package-local contracts and Go standard-library types.
package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeDirPrecedence(t *testing.T) {
	cfg, err := Load(Options{Args: []string{"--runtime-dir", "/cli/app"}, Env: map[string]string{"APPARAT_RUNTIME_DIR": "/env/app"}, DefaultMode: ModeGUI, BinaryName: "apparat"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RootDir != filepath.Clean("/cli/app") {
		t.Fatalf("root = %q", cfg.RootDir)
	}
}

func TestRuntimeDirUsesEnvironment(t *testing.T) {
	cfg, err := Load(Options{Env: map[string]string{"APPARAT_RUNTIME_DIR": "/env/app"}, DefaultMode: ModeHeadless, BinaryName: "apparatd"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RootDir != filepath.Clean("/env/app") {
		t.Fatalf("root = %q", cfg.RootDir)
	}
}

func TestDefaultRootAvoidsSourceDirectoryName(t *testing.T) {
	cfg, err := Load(Options{Env: map[string]string{"XDG_DATA_HOME": "/tmp/data"}, DefaultMode: ModeGUI, BinaryName: "apparat"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(cfg.RootDir, filepath.Join("data", "apparat", "node")) {
		t.Fatalf("root = %q", cfg.RootDir)
	}
}

func TestDefaultRootsAreShared(t *testing.T) {
	gui, err := Load(Options{Env: map[string]string{"XDG_DATA_HOME": "/tmp/data"}, DefaultMode: ModeGUI, BinaryName: "apparat"})
	if err != nil {
		t.Fatal(err)
	}
	headless, err := Load(Options{Env: map[string]string{"XDG_DATA_HOME": "/tmp/data"}, DefaultMode: ModeHeadless, BinaryName: "apparatd"})
	if err != nil {
		t.Fatal(err)
	}
	if gui.RootDir != headless.RootDir {
		t.Fatalf("expected shared roots, got = %q", gui.RootDir)
	}
}
