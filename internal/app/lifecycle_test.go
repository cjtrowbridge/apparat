// Injects startup failures and checks shared state through actual adapters.
// Exercises cleanup, ownership, REST readiness and three-service restart behavior.
package app_test

import (
	"context"
	"encoding/json"
	"github.com/cjtrowbridge/apparat/internal/adapters/mockprovider"
	"github.com/cjtrowbridge/apparat/internal/app"
	"github.com/cjtrowbridge/apparat/internal/config"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"net/http"
	"testing"
)

func nodeFor(t *testing.T, root string, demo bool) *app.Runtime {
	t.Helper()
	cfg, err := config.Load(config.Options{Args: []string{"--runtime-dir", root}, DefaultMode: config.ModeHeadless})
	if err != nil {
		t.Fatal(err)
	}
	cfg.Listen = "127.0.0.1:0"
	cfg.Demo = demo
	node, err := app.NewRuntimeWithConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	node.SetDrivers(mockprovider.Drivers())
	t.Cleanup(func() {
		_ = node.Close()
	})
	return node
}
func TestThreeServicesRESTAndRestart(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	node := nodeFor(t, root, true)
	{
		err := node.Initialize(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}
	list, err := node.List(ctx)
	if err != nil || len(list) != 3 {
		t.Fatal(list, err)
	}
	list[0].Fail = true
	{
		err = node.UpdateService(ctx, list[0])
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		err = node.Refresh(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}
	list, err = node.List(ctx)
	if err != nil || list[0].Health != "unhealthy" || list[1].Health != "healthy" || list[2].Health != "healthy" {
		t.Fatal(list, err)
	}
	response, err := http.Get("http://" + node.Address() + "/v1/services")
	if err != nil {
		t.Fatal(err)
	}
	var remote []domain.Service
	err = json.NewDecoder(response.Body).Decode(&remote)
	_ = response.Body.Close()
	if err != nil || len(remote) != 3 || remote[0].ID != list[0].ID || remote[0].Fail {
		t.Fatal(remote, err)
	}
	contender := nodeFor(t, root, false)
	if contender.Initialize(ctx) == nil {
		t.Fatal("second node acquired root")
	}
	{
		err = node.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	again := nodeFor(t, root, false)
	{
		err = again.Initialize(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}
	restored, err := again.List(ctx)
	if err != nil || len(restored) != 3 || restored[0].ID != list[0].ID {
		t.Fatal(restored, err)
	}
	{
		err = again.RemoveService(ctx, restored[0].ID)
		if err != nil {
			t.Fatal(err)
		}
	}
	restored, err = again.List(ctx)
	if err != nil || len(restored) != 2 {
		t.Fatal(restored, err)
	}
}
