package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeDirPrecedence(t *testing.T) {
	cfg, err := Load(Options{Args: []string{"--runtime-dir", "/cli/app"}, Env: map[string]string{"APPARAT_RUNTIME_DIR": "/env/app"}, DefaultMode: ModeGUI})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RootDir != filepath.Clean("/cli/app") {
		t.Fatalf("root = %q", cfg.RootDir)
	}
}

func TestRuntimeDirUsesEnvironment(t *testing.T) {
	cfg, err := Load(Options{Env: map[string]string{"APPARAT_RUNTIME_DIR": "/env/app"}, DefaultMode: ModeHeadless})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RootDir != filepath.Clean("/env/app") {
		t.Fatalf("root = %q", cfg.RootDir)
	}
}

func TestDefaultRootAvoidsSourceDirectoryName(t *testing.T) {
	cfg, err := Load(Options{Env: map[string]string{"XDG_DATA_HOME": "/tmp/data"}, DefaultMode: ModeGUI})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(cfg.RootDir, filepath.Join("data", "apparat")) {
		t.Fatalf("root = %q", cfg.RootDir)
	}
}
