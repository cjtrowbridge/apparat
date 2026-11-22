// Verifies every initialized startup stage unwinds resources before returning.
// Lifecycle injection stays test-local and does not replace real lock/database checks.
package app

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/config"
	"github.com/cjtrowbridge/apparat/internal/platform"
	"sync"
	"testing"
)

func TestStartupFailureReleasesOwnership(t *testing.T) {
	for _, stage := range []string{"lock", "directories_ready", "database", "migrations_ready", "identity", "repository_ready", "listener", "supervision", "ready"} {
		t.Run(stage, func(t *testing.T) {
			cfg, err := config.Load(config.Options{Args: []string{"--runtime-dir", t.TempDir()}, DefaultMode: config.ModeHeadless})
			if err != nil {
				t.Fatal(err)
			}
			node, err := NewRuntimeWithConfig(cfg)
			if err != nil {
				t.Fatal(err)
			}
			injected := errors.New("injected " + stage)
			node.hook = func(name string) error {
				if name == stage {
					return injected
				}
				return nil
			}
			err = node.Initialize(context.Background())
			if !errors.Is(err, injected) || node.Health().Ready || node.database != nil || node.server != nil || node.cancel != nil {
				t.Fatalf("cleanup failed: %v", err)
			}
			lock, err := platform.Acquire(cfg.RootDir, "test", "test")
			if err != nil {
				t.Fatal(err)
			}
			_ = lock.Close()
			var group sync.WaitGroup
			for i := 0; i < 3; i++ {
				group.Go(func() {
					_ = node.Close()
				})
			}
			group.Wait()
		})
	}
}

func TestShutdownOrder(t *testing.T) {
	cfg, err := config.Load(config.Options{Args: []string{"--runtime-dir", t.TempDir()}, DefaultMode: config.ModeHeadless})
	if err != nil {
		t.Fatal(err)
	}
	cfg.Listen = "127.0.0.1:0"
	node, err := NewRuntimeWithConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	{
		err = node.Initialize(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}
	stages := []string{}
	node.hook = func(stage string) error {
		stages = append(stages, stage)
		return nil
	}
	{
		err = node.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	expected := []string{"supervision_stopped", "listener_stopped", "database_closed", "lock_releasing"}
	if len(stages) != len(expected) {
		t.Fatal(stages)
	}
	for index := range stages {
		if stages[index] != expected[index] {
			t.Fatal(stages)
		}
	}
}
