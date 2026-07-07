package app

import (
	"context"
	"testing"

	"github.com/cjtrowbridge/apparat/internal/config"
)

func TestRuntimeSnapshotUsesCanonicalHUD(t *testing.T) {
	runtime, err := NewRuntime(ModeGUI)
	if err != nil {
		t.Fatal(err)
	}
	tabs := runtime.Snapshot().TabTitles()
	if len(tabs) != 7 {
		t.Fatalf("tab count = %d, want 7", len(tabs))
	}
	if tabs[0] != "Comrades" {
		t.Fatalf("first tab = %q, want Comrades", tabs[0])
	}
	if tabs[2] != "Research" {
		t.Fatalf("third tab = %q, want Research", tabs[2])
	}
}

func TestRuntimeRejectsUnknownMode(t *testing.T) {
	if _, err := NewRuntime(Mode("bogus")); err == nil {
		t.Fatal("expected unsupported mode error")
	}
}

func TestRuntimeDoctorInitializesLocalState(t *testing.T) {
	cfg, err := config.Load(config.Options{Args: []string{"--runtime-dir", t.TempDir()}, DefaultMode: config.ModeHeadless})
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := NewRuntimeWithConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Close() }()
	diag := runtime.Doctor(context.Background())
	if !diag.Healthy {
		t.Fatalf("doctor unhealthy: %+v", diag)
	}
}
