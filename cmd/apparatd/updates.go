// Runs release checks serially after the headless listener or owner signals eligibility.
// A successful promotion requests a clean node shutdown and process replacement.
package main

import (
	"context"
	"os"
	"runtime/debug"

	"github.com/cjtrowbridge/apparat/internal/adapters/discovery"
	"github.com/cjtrowbridge/apparat/internal/adapters/selfupdate"
	"github.com/cjtrowbridge/apparat/internal/app"
)

type updateEvents struct {
	auto   chan struct{}
	manual chan struct{}
}

func (events *updateEvents) Write(data []byte) (int, error) {
	for _, event := range data {
		target := events.auto
		switch event {
		case 'M':
			target = events.manual
		case 'A':
		default:
			continue
		}
		select {
		case target <- struct{}{}:
		default:
		}
	}
	return len(data), nil
}

func (events *updateEvents) run(ctx context.Context, node *app.Runtime) int {
	for {
		mode := "auto"
		select {
		case <-ctx.Done():
			return 0
		case <-events.manual:
			mode = "manual"
		case <-events.auto:
		}
		path := node.Config().DatabasePath
		result, err := checkWorker(ctx, func(stage string) bool {
			return discovery.Command(ctx, "updates-"+stage, path, mode).Run() == nil
		})
		level := "info"
		if err != nil {
			level = "error"
			result = err.Error()
		}
		record := discovery.Command(ctx, "updates-result", path, result)
		if record.Run() != nil && ctx.Err() == nil {
			_ = node.RecordLastRun("error", "updater", "status_write_failed", "Could not save update result", nil)
		}
		_ = node.RecordLastRun(level, "updater", "worker_check", result, nil)
		if result == "installed" {
			return 75
		}
	}
}

func checkWorker(ctx context.Context, eligible func(string) bool) (string, error) {
	if !eligible("eligible") {
		return "deferred", nil
	}
	client := selfupdate.NewWorkerClient()
	release, err := client.Latest(ctx)
	if err != nil {
		return "", err
	}
	target, err := os.Executable()
	if err != nil {
		return "", selfupdate.ErrInstall
	}
	digest, err := selfupdate.Digest("/proc/self/exe")
	if err != nil {
		return "", err
	}
	if digest == release.SHA256 {
		return "current", nil
	}
	var revision string
	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				revision = setting.Value
			}
		}
	}
	err = client.CheckSuccessor(ctx, revision, release.Revision)
	if err != nil {
		return "", err
	}
	candidate, err := client.Download(ctx, release, target)
	if err != nil {
		return "", err
	}
	defer candidate.Discard()
	if candidate.Original != digest {
		return "", selfupdate.ErrVerification
	}
	if !eligible("drain") {
		return "deferred", nil
	}
	err = candidate.Install(ctx)
	if err != nil {
		return "", err
	}
	return "installed", nil
}
