//go:build gui

// Exercises the Cluster inventory selectors, port form and device detail on both layouts.
// APPARAT_INVENTORY_PNG captures one selected layout through the real display renderer.
package gui

import (
	"image"
	"image/png"
	"os"
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
	if strings.Contains(path, "narrow") {
		w, h = 390, 844
	}
	g := inventoryGame(t, w, h)
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
