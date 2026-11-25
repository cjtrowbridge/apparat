//go:build gui

// Exercises the Cluster inventory selectors, port form and device detail on both layouts.
// APPARAT_INVENTORY_PNG captures one selected layout through the real display renderer.
package gui

import (
	"context"
	"github.com/ebitenui/ebitenui/event"
	"image"
	"image/png"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

func inventoryGame(t *testing.T, width, height int) *Game {
	t.Helper()
	g := horizontalGame()
	g.Layout(width, height)
	g.runtimeInfo.Queries = &projectionSource{}
	v := domain.WorkerInventory{Observed: time.Now().Unix(), Version: "0.14.1", Port: 11434, Models: []domain.InventoryModel{{Name: "trashcan0-local-256k-qwen3.8-flash-next:125b-a6b-q4_K_M", Text: true, Loaded: true}}, Resources: domain.NodeResources{OS: "linux", Arch: "amd64", CPUs: 12, RAM: &domain.Capacity{Total: 32 << 30, Used: 8 << 30, Available: 24 << 30}}}
	g.serviceProjection.Devices = []domain.ClusterDevice{{Fingerprint: "fixture", Host: "trashcan0", Admin: true, Inventory: &v}}
	selectTabThroughStrip(t, g, hud.TabCluster)
	tab := g.presentedTab(g.shell.Snapshot().ActiveTab())
	selected := -1
	icons := 0
	for i, s := range tab.Sections {
		if s.CenterIcon != "" {
			icons++
		}
		if s.ActionID == "worker-fixture" {
			selected = i
			if s.CapabilityIcon == "" || len(s.DetailSections) != 6 {
				t.Fatal("missing device inventory or badge")
			}
			pane := widget.NewContainer()
			g.sectionDetail(pane, hud.TabCluster, s)
			if len(pane.Children()) < 7 {
				t.Fatal("device details must render as seven sibling cards")
			}
			for _, child := range pane.Children()[:7] {
				_, card := child.(*widget.Container)
				if !card {
					t.Fatal("expected a separate card")
				}
			}
			text := s.DetailSections[1]
			if text.ActionID != "" || len(text.Rows) != 1 || !strings.Contains(text.Rows[0].Detail, "Loaded") {
				t.Fatal(text)
			}
		}
	}
	if selected < 0 || icons != 5 {
		t.Fatal(selected, icons)
	}
	g.selectSection(hud.TabCluster, selected)
	renderNavigation(g)
	return g
}

func TestInventoryDeviceView(t *testing.T) {
	for _, size := range []image.Point{{1280, 800}, {390, 844}} {
		g := inventoryGame(t, size.X, size.Y)
		v := g.serviceProjection.Devices[0].Inventory
		v.Models = nil
		checkPort := func(want bool) {
			t.Helper()
			card := inventoryDetail(v, "fixture")[1]
			found := false
			for _, o := range g.deviceManagementLinks("worker-fixture") {
				if o.Route.Page == "ollama" {
					found = true
				}
			}
			if found != want || (card.ActionID != "") != want {
				t.Fatal("port control visibility", card, found, want)
			}
		}
		checkPort(false)
		v.Problem = "service unavailable"
		checkPort(true)
		v.Problem = ""
		v.Observed -= 121
		g.serviceProjection.Devices[0].Admin = false
		tab := g.presentedTab(g.shell.Snapshot().ActiveTab())
		for _, s := range tab.Sections {
			if s.ActionID == "worker-fixture" && (s.CapabilityIcon != "" || s.DetailSections[1].ActionID != "" || !strings.Contains(s.DetailSections[0].Description, "stale")) {
				t.Fatal("stale or unauthorized inventory controls", s)
			}
		}
	}
}

type inventoryCapture struct {
	g    *Game
	path string
	done bool
	err  error
}

func (c *inventoryCapture) Update() error {
	if c.done {
		return ebiten.Termination
	}
	return nil
}
func (c *inventoryCapture) Layout(int, int) (int, int) {
	return c.g.width, c.g.height
}
func (c *inventoryCapture) Draw(dst *ebiten.Image) {
	if c.done {
		return
	}
	c.g.ui.Container.SetLocation(dst.Bounds())
	c.g.ui.Draw(dst)
	frame := image.NewRGBA(dst.Bounds())
	dst.ReadPixels(frame.Pix)
	f, err := os.Create(c.path)
	if err == nil {
		err = png.Encode(f, frame)
		closeErr := f.Close()
		if err == nil {
			err = closeErr
		}
	}
	c.err = err
	c.done = true
}

func TestCaptureInventoryDevice(t *testing.T) {
	path := os.Getenv("APPARAT_INVENTORY_PNG")
	if path == "" {
		t.Skip("optional rendered evidence")
	}
	w, h := 1280, 800
	if strings.Contains(path, "1080") {
		w, h = 1920, 1080
	}
	if strings.Contains(path, "narrow") {
		w, h = 390, 844
	}
	g := inventoryGame(t, w, h)
	if strings.Contains(path, "hardware") {
		v := g.serviceProjection.Devices[0].Inventory
		v.Resources.PhysicalCores = 6
		v.Resources.CPUFP32PerSecond = 56e9
		v.Resources.GPUComputeReports = 2
		v.Resources.GPUFP32PerSecond = 7e12
		v.Resources.GPUs = []domain.Capacity{{Total: 6 << 30}, {Total: 6 << 30}}
		v.Resources.Disks = []domain.Capacity{{Total: 1 << 40}}
		for i, s := range g.presentedTab(g.shell.Snapshot().ActiveTab()).Sections {
			if s.ActionID == "cluster-status" {
				g.selectSection(hud.TabCluster, i)
				renderNavigation(g)
				break
			}
		}
	}
	if strings.Contains(path, "models") {
		g.verticalScrolls[len(g.verticalScrolls)-1].ScrollTop = 0.72
	}
	ebiten.SetWindowSize(w, h)
	c := &inventoryCapture{g: g, path: path}
	err := ebiten.RunGame(c)
	if err != nil || c.err != nil {
		t.Fatal(err, c.err)
	}
}

func TestClusterStatusCards(t *testing.T) {
	for _, size := range []image.Point{{390, 844}, {1280, 800}} {
		g := inventoryGame(t, size.X, size.Y)
		now := time.Now().UnixMilli()
		d := &g.serviceProjection.Devices[0]
		d.Inference = &domain.InferenceStatus{Queued: 1, Validators: 2, BlockedSince: now, Timing: domain.AgreementTiming{Recheck: 3 * time.Second}}
		tab := g.presentedTab(g.shell.Snapshot().ActiveTab())
		found := false
		for i, s := range tab.Sections {
			if s.ActionID != "cluster-status" {
				continue
			}
			found = true
			g.selectSection(hud.TabCluster, i)
			renderNavigation(g)
			if len(s.DetailSections) != 5 || s.DetailSections[0].Title != "Cluster hardware" || !strings.Contains(s.DetailSections[4].Rows[0].Detail, "idle with eligible work") {
				t.Fatal("status did not explain stalled queue")
			}
			var gear *widget.Button
			for _, target := range g.bodyTargets() {
				if focusLabel(target.w) == "Cluster settings" {
					gear, _ = target.w.(*widget.Button)
				}
			}
			if gear == nil {
				t.Fatal("missing status settings button", size)
			}
			gear.Click()
			event.ExecuteDeferred()
			if g.clusterTop() == nil || g.clusterTop().route.Page != "settings" {
				t.Fatal("gear did not open settings")
			}
			g.clusterBack()
			links := g.deviceManagementLinks(s.ActionID)
			if len(links) != 0 {
				t.Fatal("unexpected manual inference setup")
			}
		}
		if !found {
			t.Fatal("missing cluster status destination")
		}
	}
}

func TestClusterHardwareTotals(t *testing.T) {
	now := time.Unix(1000, 0)
	v := domain.WorkerInventory{Observed: 1000, Resources: domain.NodeResources{HostBootDigest: "same-host", CPUs: 16, PhysicalCores: 8, RAM: &domain.Capacity{Total: 32 << 30}, GPUs: []domain.Capacity{{Total: 6 << 30}, {Total: 6 << 30}}, Disks: []domain.Capacity{{Total: 2 << 40}}, GPUComputeReports: 2, GPUFP32PerSecond: 1e15, CPUFP32PerSecond: 9e14}}
	other := domain.WorkerInventory{Observed: 995, Resources: domain.NodeResources{CPUs: 8, CPUFP32PerSecond: 1e14}}
	stale := v
	stale.Observed = 800
	p := ServiceProjection{LocalDevice: domain.PoolDevice{ID: "gui"}, LocalInventory: v, Devices: []domain.ClusterDevice{{Fingerprint: "daemon", Inventory: &v}, {Fingerprint: "other", Inventory: &other}, {Fingerprint: "old", Inventory: &stale}}}
	s := hardwareCard(p, now)
	want := []hud.Row{{Label: "CPU cores (1/2)", Detail: "8"}, {Label: "CPU threads", Detail: "24"}, {Label: "RAM (1/2)", Detail: "32 GiB"}, {Label: "GPU VRAM (1/2)", Detail: "12 GiB"}, {Label: "Disk capacity (1/2)", Detail: "2 TiB"}, {Label: "Est. INT8 compute", Detail: "8 POP/s"}}
	if !slices.Equal(s.Rows, want) || !strings.HasPrefix(s.Description, "2 reporting hosts") {
		t.Fatal(s)
	}
	if hardwareUnits(2e12, 1000, []string{"FLOP/s", "kFLOP/s", "MFLOP/s", "GFLOP/s", "TFLOP/s"}) != "2 TFLOP/s" {
		t.Fatal("tera scaling")
	}
}

func TestHardwareMembershipChanges(t *testing.T) {
	now := time.Now()
	gui := domain.HardwareCapacity{Boot: "gui", Threads: 4}
	worker := domain.HardwareCapacity{Boot: "worker", Threads: 40}
	p := ServiceProjection{LocalDevice: domain.PoolDevice{ID: "local"}, LocalInventory: domain.WorkerInventory{Observed: now.Unix(), Resources: domain.NodeResources{CPUs: 8}}}
	threads := func(want string) {
		t.Helper()
		card := hardwareCard(p, now)
		if len(card.Rows) < 2 || card.Rows[1].Detail != want {
			t.Fatalf("want %s threads: %+v", want, card)
		}
	}
	threads("8")
	p.Devices = []domain.ClusterDevice{{Fingerprint: "worker", Hardware: map[string]domain.HardwareCapacity{"worker": worker, "gui": gui}, Roster: []domain.PoolDevice{{ID: "gui"}}}}
	threads("52")
	p.Devices[0].RosterHardware = map[string]domain.HardwareCapacity{"worker": worker, "other": {Boot: "other", Threads: 24}}
	threads("76")
	p.Devices[0].Roster = nil
	threads("72")
	p.Devices = nil
	threads("8")
}

func TestInterruptedWorkerPresentation(t *testing.T) {
	for _, size := range []image.Point{{390, 844}, {1280, 800}} {
		g := inventoryGame(t, size.X, size.Y)
		d := &g.serviceProjection.Devices[0]
		v := d.Inventory
		v.LastContact = time.Now().Add(-3 * time.Minute).Unix()
		v.Observed = v.LastContact
		v.Retrying = true
		tab := g.presentedTab(g.shell.Snapshot().ActiveTab())
		for _, s := range tab.Sections {
			switch s.ActionID {
			case "worker-fixture":
				if s.CapabilityIcon == "" || len(s.DetailSections[1].Rows) != 1 || !strings.Contains(s.Description, "Retrying automatically") {
					t.Fatal("lost worker state", s)
				}
			case "text-generation":
				if len(s.Rows) != 1 || !strings.Contains(s.Rows[0].Detail, "Last heard from trashcan0") {
					t.Fatal("lost service advertisement", s)
				}
			}
		}
		catalog := domain.ChatModelCatalog(domain.WorkerInventory{}, domain.PoolDevice{}, g.serviceProjection.Devices, time.Now())
		if len(catalog) != 1 || v.HasText(time.Now()) {
			t.Fatal("catalog and live readiness diverged from expected state")
		}
		g.clusterPush(clusterFrame{route: domain.ClusterRoute{Page: "ollama", Worker: d.Fingerprint}})
		g.cluster.pending = make(chan clusterResponse, 1)
		g.cluster.pending <- clusterResponse{err: context.DeadlineExceeded}
		g.pollCluster()
		if g.clusterTop().retryAt.IsZero() || g.clusterTop().notice != "" {
			t.Fatal("failed read did not schedule retry")
		}
		renderNavigation(g)
		g.cluster.pending = make(chan clusterResponse, 1)
		g.cluster.pending <- clusterResponse{page: domain.ClusterPage{Title: "Configure Ollama"}}
		g.pollCluster()
		if !g.clusterTop().retryAt.IsZero() || g.clusterTop().page.Title != "Configure Ollama" {
			t.Fatal("failed to recover detail")
		}
	}
}

func TestSharedTaskStatus(t *testing.T) {
	receipt := domain.TaskObservation{ID: "turn", Turn: "turn", Pool: "Chat Pool", State: "received"}
	queued := domain.TaskObservation{ID: "job", Turn: "turn", Pool: "Chat Pool", State: "queued"}
	done := queued
	done.State = "finished"
	card := taskStatusCard([]domain.ClusterDevice{
		{Inference: &domain.InferenceStatus{Recent: []domain.TaskObservation{receipt, queued}}},
		{Inference: &domain.InferenceStatus{Recent: []domain.TaskObservation{done}}},
	})
	if len(card.Rows) != 1 || card.Rows[0].Detail != "job" || !strings.Contains(card.Rows[0].Label, "finished") {
		t.Fatal(card)
	}
}
