// Verifies initializes and coordinates shared local application runtime resources.
// Dependencies: github.com/cjtrowbridge/apparat/internal/config.
package app

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/config"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

func TestRuntimeRejectsUnknownMode(t *testing.T) {
	{
		_, err := NewRuntime(Mode("bogus"))
		if err == nil {
			t.Fatal("expected unsupported mode error")
		}
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
	defer func() {
		_ = runtime.Close()
	}()
	diag := runtime.Doctor(context.Background())
	if !diag.Healthy {
		t.Fatalf("doctor unhealthy: %+v", diag)
	}
	if diag.LastRunPath != cfg.LastRunPath {
		t.Fatalf("doctor last run path = %q, want %q", diag.LastRunPath, cfg.LastRunPath)
	}
	data, err := os.ReadFile(cfg.LastRunPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"process_start", "last_run", "directories_ready", "migrations_ready", "repository_ready", "healthy"} {
		if !strings.Contains(text, want) {
			t.Fatalf("last_run missing %q: %s", want, text)
		}
	}
}

func TestRuntimeDefaultsUseSharedNodeRoot(t *testing.T) {
	root := t.TempDir()
	guiCfg, err := config.Load(config.Options{Env: map[string]string{"XDG_DATA_HOME": root}, DefaultMode: config.ModeGUI, BinaryName: "apparat"})
	if err != nil {
		t.Fatal(err)
	}
	headlessCfg, err := config.Load(config.Options{Env: map[string]string{"XDG_DATA_HOME": root}, DefaultMode: config.ModeHeadless, BinaryName: "apparatd"})
	if err != nil {
		t.Fatal(err)
	}
	if guiCfg.LastRunPath != headlessCfg.LastRunPath {
		t.Fatalf("expected shared last_run path, got = %q", guiCfg.LastRunPath)
	}
}

func TestRenameLocalWithoutWorker(t *testing.T) {
	n := Runtime{Keys: identity.NewSession(t.TempDir()), Personal: identity.NewPersonal(t.TempDir())}
	err := n.Keys.Install(nil, "", "plain", "", "", true)
	if err != nil {
		t.Fatal(err)
	}
	err = n.Personal.Create("Owner", []byte("test password"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	route := domain.ClusterRoute{Page: "rename-local", Target: "some-other-fingerprint"}
	_, err = n.ClusterPage(ctx, route)
	if err == nil {
		t.Fatal("wrong device accepted")
	}
	route.Target = n.LocalClusterDevice().ID
	page, err := n.ClusterPage(ctx, route)
	if err != nil || len(page.Options) != 1 {
		t.Fatal("local rename unavailable", err)
	}
	err = page.Options[0].Run(ctx, "Pixel 10 XL")
	if err != nil || n.LocalClusterDevice().Host != "Pixel-10-XL" {
		t.Fatal("local rename failed", err)
	}
}

func TestEffectivePoolUserList(t *testing.T) {
	v := persistence.GovernanceView{Memberships: []persistence.GroupMember{{User: "me", Name: "team"}}}
	member := persistence.Member{ID: "me", Admission: "approved"}
	for _, target := range []string{"public", "me", "group:team", "other", "group:other"} {
		v.Grants = []persistence.Membership{{Pool: "Chat", Service: persistence.HashService, Principal: target}}
		expected := target == "public" || target == "me" || target == "group:team"
		if poolAuthorized(&v, member, "Chat") != expected || poolAuthorized(&v, member, "Other") {
			t.Fatal("wrong effective permission", target)
		}
	}
	member.Admin = true
	if !poolAuthorized(&v, member, "Other") {
		t.Fatal("missing administrator")
	}
	member.Admission = "blocked"
	if poolAuthorized(&v, member, "Chat") {
		t.Fatal("included blocked account")
	}
}
